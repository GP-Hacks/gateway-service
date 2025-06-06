package auth

import (
	"log/slog"
	"net/http"

	common "github.com/GP-Hacks/kdt2024-commons/json"
	proto "github.com/GP-Hacks/proto/pkg/api/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type signInReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewSignInHandler(log *slog.Logger, authClient proto.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		select {
		case <-ctx.Done():
			common.WriteError(w, http.StatusRequestTimeout, "Request timed out")
			return
		default:
		}

		var reqJ signInReq
		if err := common.ReadJSON(r, &reqJ); err != nil {
			common.WriteError(w, http.StatusBadRequest, "Invalid body")
			return
		}

		req := &proto.SignInRequest{
			Email:    reqJ.Email,
			Password: reqJ.Password,
		}

		resp, err := authClient.SignIn(ctx, req)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				common.WriteError(w, http.StatusNotFound, "User not found")
				return
			}
			if status.Code(err) == codes.Unauthenticated {
				common.WriteError(w, http.StatusUnauthorized, "Invalid credentials")
				return
			}
			log.Debug("failed sign in: %v", err)
			common.WriteError(w, http.StatusInternalServerError, "")
			return
		}

		common.WriteJSON(w, http.StatusOK, resp.Tokens)
	}
}
