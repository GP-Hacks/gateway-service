package chat

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/GP-Hacks/kdt2024-commons/json"
	"github.com/GP-Hacks/kdt2024-gateway/internal/utils"
	proto "github.com/GP-Hacks/proto/pkg/api/chat"
	"github.com/go-chi/chi/v5/middleware"
)

type historyMsg struct {
	Content   string    `json:"content,omitempty"`
	Role      string    `json:"role,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

func NewGetHistoryHandler(log *slog.Logger, chatClient proto.ChatServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.chat.send.New"
		ctx := r.Context()
		reqID := middleware.GetReqID(ctx)
		logger := log.With(
			slog.String("operation", op),
			slog.String("request_id", reqID),
			slog.String("client_ip", r.RemoteAddr),
			slog.String("method", r.Method),
			slog.String("url", r.URL.String()),
		)

		logger.Info("Handling request to send message")

		select {
		case <-ctx.Done():
			logger.Warn("Request was cancelled by the client")
			http.Error(w, "Request was cancelled", http.StatusRequestTimeout)
			return
		default:
		}

		token, err := utils.GetTokenFromHeader(r)
		if err != nil {
			json.WriteError(w, http.StatusUnauthorized, "Invalid authorization header")
			return
		}

		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")

		limitI, err := strconv.Atoi(limit)
		if err != nil {
			json.WriteError(w, http.StatusBadRequest, "Invalid query")
			return
		}

		offsetI, err := strconv.Atoi(offset)
		if err != nil {
			json.WriteError(w, http.StatusBadRequest, "Invalid query")
			return
		}

		res, err := chatClient.GetHistory(ctx, &proto.GetHistoryRequest{
			Token:  token,
			Limit:  int64(limitI),
			Offset: int64(offsetI),
		})
		if err != nil {
			json.WriteError(w, http.StatusInternalServerError, "")
			return
		}

		resp := make([]historyMsg, len(res.Messages))
		for i, m := range res.Messages {
			var role string
			if m.Role == proto.ChatRole_BOT {
				role = "bot"
			} else {
				role = "user"
			}

			resp[i] = historyMsg{
				Role:      role,
				Content:   m.Content,
				CreatedAt: m.CreatedAt.AsTime(),
			}
		}

		logger.Debug("Message sent successfully", slog.Any("response", resp))
		json.WriteJSON(w, http.StatusOK, resp)
	}
}
