package auth

import (
	"net/http"

	common "github.com/GP-Hacks/kdt2024-commons/json"
	proto "github.com/GP-Hacks/proto/pkg/api/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type refreshTokensReq struct {
	RefreshToken string `json:"refresh_token"`
}

func NewRefreshTokensHandler(authClient proto.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		select {
		case <-ctx.Done():
			common.WriteError(w, http.StatusRequestTimeout, "Request timed out")
			return
		default:
		}

		var reqJ refreshTokensReq
		if err := common.ReadJSON(r, &reqJ); err != nil {
			common.WriteError(w, http.StatusBadRequest, "Invalid body")
			return
		}

		req := &proto.RefreshTokensRequest{
			RefreshToken: reqJ.RefreshToken,
		}

		resp, err := authClient.RefreshTokens(ctx, req)
		if err != nil {
			if status.Code(err) == codes.Unauthenticated {
				common.WriteError(w, http.StatusUnauthorized, "Invalid refresh token")
				return
			}
			common.WriteError(w, http.StatusInternalServerError, "")
			return
		}

		common.WriteJSON(w, http.StatusOK, resp.Tokens)
	}
}
