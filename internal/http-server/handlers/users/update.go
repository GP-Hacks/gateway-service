package users

import (
	"log"
	"net/http"
	"time"

	common "github.com/GP-Hacks/kdt2024-commons/json"
	"github.com/GP-Hacks/kdt2024-gateway/internal/utils"
	proto "github.com/GP-Hacks/proto/pkg/api/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type updateUserReq struct {
	User struct {
		Email       string    `json:"email"`
		FirstName   string    `json:"first_name"`
		LastName    string    `json:"last_name"`
		Surname     string    `json:"surname"`
		DateOfBirth time.Time `json:"date_of_birth"`
	} `json:"user"`
}

func NewUpdateHandler(userClient proto.UserServiceClient) http.HandlerFunc {
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

		var reqJ updateUserReq
		if err := common.ReadJSON(r, &reqJ); err != nil {
			common.WriteError(w, http.StatusBadRequest, "Invalid body")
			return
		}

		log.Print(r)
		log.Print(reqJ)

		req := &proto.UpdateUserRequest{
			Token: token,
			User: &proto.User{
				Email:       reqJ.User.Email,
				FirstName:   reqJ.User.FirstName,
				LastName:    reqJ.User.LastName,
				Surname:     reqJ.User.Surname,
				DateOfBirth: timestamppb.New(reqJ.User.DateOfBirth),
			},
		}

		log.Print(req)

		_, err = userClient.Update(ctx, req)
		if err != nil {
			if status.Code(err) == codes.Unauthenticated {
				common.WriteError(w, http.StatusUnauthorized, "Invalid token")
				return
			}
			if status.Code(err) == codes.NotFound {
				common.WriteError(w, http.StatusNotFound, "User not found")
				return
			}
			log.Print(err)
			common.WriteError(w, http.StatusInternalServerError, "")
			return
		}

		common.WriteJSON(w, http.StatusOK, "")
	}
}
