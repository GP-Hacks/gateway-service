package auth

import (
	"net/http"
	"time"

	common "github.com/GP-Hacks/kdt2024-commons/json"
	proto "github.com/GP-Hacks/proto/pkg/api/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type signUpReq struct {
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Surname     string    `json:"surname"`
	Password    string    `json:"password"`
	DateOfBirth time.Time `json:"date_of_birth"`
}

func NewSignUpHandler(authClient proto.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handler.charity.getcategories.New"
		ctx := r.Context()
		select {
		case <-ctx.Done():
			// logger.Warn("Request cancelled by the client")
			common.WriteError(w, http.StatusRequestTimeout, "Request timed out")
			return
		default:
		}

		var reqJ signUpReq
		if err := common.ReadJSON(r, &reqJ); err != nil {
			common.WriteError(w, http.StatusBadRequest, "Invalid body")
			return
		}

		req := &proto.SignUpRequest{
			Email:       reqJ.Email,
			FirstName:   reqJ.FirstName,
			LastName:    reqJ.LastName,
			Surname:     reqJ.Surname,
			Password:    reqJ.Password,
			DateOfBirth: timestamppb.New(reqJ.DateOfBirth),
		}
		// logger.Debug("Sending request to auth service", slog.Any("request", req))

		_, err := authClient.SignUp(ctx, req)
		if err != nil {
			if status.Code(err) == codes.AlreadyExists {
				// logger.Warn("User already exists", slog.String("error", err.Error()))
				common.WriteError(w, http.StatusAlreadyReported, "User already exists")
				return
			}
			common.WriteError(w, http.StatusInternalServerError, "")
			return
		}

		common.WriteJSON(w, http.StatusOK, "")
	}
}
