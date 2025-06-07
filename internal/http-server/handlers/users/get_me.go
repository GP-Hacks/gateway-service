package users

import (
	"net/http"

	common "github.com/GP-Hacks/kdt2024-commons/json"
	"github.com/GP-Hacks/kdt2024-gateway/internal/utils"
	proto "github.com/GP-Hacks/proto/pkg/api/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewGetMeHandler(userClient proto.UserServiceClient) http.HandlerFunc {
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

		req := &proto.GetMeRequest{
			Token: token,
		}

		resp, err := userClient.GetMe(ctx, req)
		if err != nil {
			if status.Code(err) == codes.Unauthenticated {
				common.WriteError(w, http.StatusUnauthorized, "Invalid token")
				return
			}
			if status.Code(err) == codes.NotFound {
				common.WriteError(w, http.StatusNotFound, "User not found")
				return
			}
			common.WriteError(w, http.StatusInternalServerError, "")
			return
		}

		response := map[string]interface{}{
			"id":            resp.Id,
			"email":         resp.User.Email,
			"first_name":    resp.User.FirstName,
			"last_name":     resp.User.LastName,
			"surname":       resp.User.Surname,
			"date_of_birth": resp.User.DateOfBirth.AsTime(),
			"avatar_url":    resp.AvatarURL,
			"status":        resp.Status,
		}

		common.WriteJSON(w, http.StatusOK, response)
	}
}
