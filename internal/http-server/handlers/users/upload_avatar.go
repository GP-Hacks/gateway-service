package users

import (
	"net/http"

	common "github.com/GP-Hacks/kdt2024-commons/json"
	"github.com/GP-Hacks/kdt2024-gateway/internal/utils"
	proto "github.com/GP-Hacks/proto/pkg/api/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type uploadAvatarReq struct {
	Photo []byte `json:"photo"`
}

func NewUploadAvatarHandler(userClient proto.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		select {
		case <-ctx.Done():
			common.WriteError(w, http.StatusRequestTimeout, "Request timed out")
			return
		default:
		}

		token, err := utils.GetTokenFromHeader(r)
		if err != nil {
			common.WriteError(w, http.StatusUnauthorized, "Invalid authorization header")
			return
		}

		var reqJ uploadAvatarReq
		if err := common.ReadJSON(r, &reqJ); err != nil {
			common.WriteError(w, http.StatusBadRequest, "Invalid body")
			return
		}

		req := &proto.UploadAvatarRequest{
			Photo: reqJ.Photo,
			Token: token,
		}

		resp, err := userClient.UploadAvatar(ctx, req)
		if err != nil {
			if status.Code(err) == codes.Unauthenticated {
				common.WriteError(w, http.StatusUnauthorized, "Invalid token")
				return
			}
			common.WriteError(w, http.StatusInternalServerError, "")
			return
		}

		common.WriteJSON(w, http.StatusOK, map[string]string{"url": resp.Url})
	}
}
