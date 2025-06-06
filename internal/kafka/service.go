package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/GP-Hacks/kdt2024-gateway/config"
	"github.com/IBM/sarama"
)

type KafkaService struct {
	producer        sarama.SyncProducer
	consumer        sarama.Consumer
	requestTopic    string
	responseTopic   string
	pendingRequests map[string]chan []byte
	mu              sync.RWMutex
}

type RequestMessage struct {
	AuthToken string `json:"auth_token"`
	Content   string `json:"content"`
}

type SuccessResponse struct {
	Status    string `json:"status"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type ErrorResponse struct {
	Status    string `json:"status"`
	Error     string `json:"error"`
	CreatedAt string `json:"created_at"`
}

func NewKafkaService(config *config.Config) (*KafkaService, error) {
	kafkfaCfg := sarama.NewConfig()
	kafkfaCfg.Producer.Return.Successes = true
	kafkfaCfg.Consumer.Return.Errors = true

	producer, err := sarama.NewSyncProducer(config.KafkaBrokers, kafkfaCfg)
	if err != nil {
		return nil, fmt.Errorf("failed create producer: %v", err)
	}

	consumer, err := sarama.NewConsumer(config.KafkaBrokers, kafkfaCfg)
	if err != nil {
		producer.Close()
		return nil, fmt.Errorf("failed create consumer: %v", err)
	}

	return &KafkaService{
		producer:        producer,
		consumer:        consumer,
		requestTopic:    config.RequestTopic,
		responseTopic:   config.ResponseTopic,
		pendingRequests: make(map[string]chan []byte),
	}, nil
}

func (ks *KafkaService) SendMessage(uuid string, requestMsg RequestMessage) error {
	msgBytes, err := json.Marshal(requestMsg)
	if err != nil {
		return fmt.Errorf("failed serialize message: %v", err)
	}

	kafkaMsg := &sarama.ProducerMessage{
		Topic: ks.requestTopic,
		Key:   sarama.StringEncoder(uuid),
		Value: sarama.ByteEncoder(msgBytes),
	}

	_, _, err = ks.producer.SendMessage(kafkaMsg)
	if err != nil {
		return fmt.Errorf("failed send to Kafka: %v", err)
	}

	return nil
}

func (ks *KafkaService) WaitForResponse(uuid string, timeout time.Duration) ([]byte, error) {
	responseChan := make(chan []byte, 1)

	ks.mu.Lock()
	ks.pendingRequests[uuid] = responseChan
	ks.mu.Unlock()

	defer func() {
		ks.mu.Lock()
		delete(ks.pendingRequests, uuid)
		ks.mu.Unlock()
		close(responseChan)
	}()

	select {
	case response := <-responseChan:
		return response, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("ttl for UUID: %s", uuid)
	}
}

func (ks *KafkaService) StartResponseConsumer(ctx context.Context) error {
	partitionConsumer, err := ks.consumer.ConsumePartition(ks.responseTopic, 0, sarama.OffsetNewest)
	if err != nil {
		return fmt.Errorf("ошибка создания partition consumer: %v", err)
	}

	go func() {
		defer partitionConsumer.Close()

		for {
			select {
			case msg := <-partitionConsumer.Messages():
				messageUUID := string(msg.Key)

				ks.mu.RLock()
				if responseChan, exists := ks.pendingRequests[messageUUID]; exists {
					select {
					case responseChan <- msg.Value:
						log.Printf("UUID: %s", messageUUID)
					default:
						log.Printf("Канал заблокирован для UUID: %s", messageUUID)
					}
				} else {
					log.Printf("Не найден ожидающий запрос для UUID: %s", messageUUID)
				}
				ks.mu.RUnlock()

			case err := <-partitionConsumer.Errors():
				log.Printf("Ошибка консьюмера: %v", err)

			case <-ctx.Done():
				log.Println("Остановка консьюмера ответов")
				return
			}
		}
	}()

	return nil
}

func (ks *KafkaService) Close() error {
	if err := ks.producer.Close(); err != nil {
		log.Printf("Ошибка закрытия продюсера: %v", err)
	}
	if err := ks.consumer.Close(); err != nil {
		log.Printf("Ошибка закрытия консьюмера: %v", err)
	}
	return nil
}
