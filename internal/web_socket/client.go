package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/GP-Hacks/kdt2024-gateway/internal/kafka"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn       *websocket.Conn
	send       chan []byte
	processing bool
	mu         sync.Mutex
}

func (c *Client) readPump(hub *Hub, kafkaService *kafka.KafkaService, timeout time.Duration) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("Ошибка чтения WebSocket: %v", err)
			break
		}

		c.mu.Lock()
		if c.processing {
			c.mu.Unlock()
			log.Println("Клиент уже обрабатывает сообщение, игнорируем новое")
			continue
		}
		c.processing = true
		c.mu.Unlock()

		go func(msg []byte) {
			defer func() {
				c.mu.Lock()
				c.processing = false
				c.mu.Unlock()
			}()

			var requestMsg kafka.RequestMessage
			if err := json.Unmarshal(msg, &requestMsg); err != nil {
				log.Printf("Ошибка парсинга сообщения от клиента: %v", err)
				errorResponse := kafka.ErrorResponse{
					Status:    "error",
					Error:     "invalid message format",
					CreatedAt: time.Now().Format(time.RFC3339Nano),
				}
				errorBytes, _ := json.Marshal(errorResponse)
				select {
				case c.send <- errorBytes:
				default:
					log.Println("Канал отправки заблокирован")
				}
				return
			}

			messageUUID := uuid.New().String()
			if err := kafkaService.SendMessage(messageUUID, requestMsg); err != nil {
				log.Printf("Ошибка отправки в Kafka: %v", err)
				errorResponse := kafka.ErrorResponse{
					Status:    "error",
					Error:     "failed to process message",
					CreatedAt: time.Now().Format(time.RFC3339Nano),
				}
				errorBytes, _ := json.Marshal(errorResponse)
				select {
				case c.send <- errorBytes:
				default:
					log.Println("Канал отправки заблокирован")
				}
				return
			}

			response, err := kafkaService.WaitForResponse(messageUUID, timeout)
			if err != nil {
				log.Printf("Ошибка ожидания ответа: %v", err)
				errorResponse := kafka.ErrorResponse{
					Status:    "error",
					Error:     "request timeout",
					CreatedAt: time.Now().Format(time.RFC3339Nano),
				}
				errorBytes, _ := json.Marshal(errorResponse)
				select {
				case c.send <- errorBytes:
				default:
					log.Println("Канал отправки заблокирован")
				}
				return
			}

			select {
			case c.send <- response:
			default:
				log.Println("Канал отправки заблокирован")
			}
		}(message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Ошибка записи WebSocket: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ServeWS(hub *Hub, kafkaService *kafka.KafkaService, timeout time.Duration, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка upgrade WebSocket: %v", err)
		return
	}

	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}

	hub.register <- client

	go client.writePump()
	go client.readPump(hub, kafkaService, timeout)
}
