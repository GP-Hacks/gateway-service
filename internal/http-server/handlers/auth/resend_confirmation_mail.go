package auth

import (
	"net/http"

	common "github.com/GP-Hacks/kdt2024-commons/json"
	proto "github.com/GP-Hacks/proto/pkg/api/auth"
)

type resendConfirmationMailReq struct {
	Email string `json:"email"`
}

func NewResendConfiramtionMailHandler(authClient proto.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		select {
		case <-ctx.Done():
			common.WriteError(w, http.StatusRequestTimeout, "Request timed out")
			return
		default:
		}

		var reqJ resendConfirmationMailReq
		if err := common.ReadJSON(r, &reqJ); err != nil {
			common.WriteError(w, http.StatusBadRequest, "Invalid body")
			return
		}

		req := &proto.ResendConfirmationMailRequest{
			Email: reqJ.Email,
		}

		authClient.ResendConfirmationMail(ctx, req)
		common.WriteJSON(w, http.StatusOK, "")
	}
}
