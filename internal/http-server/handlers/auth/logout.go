package auth

import (
	"net/http"

	common "github.com/GP-Hacks/kdt2024-commons/json"
	proto "github.com/GP-Hacks/proto/pkg/api/auth"
)

type logoutReq struct {
	Tokens struct {
		Access  string `json:"access"`
		Refresh string `json:"refresh"`
	} `json:"tokens"`
}

func NewLogoutHandler(authClient proto.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		select {
		case <-ctx.Done():
			common.WriteError(w, http.StatusRequestTimeout, "Request timed out")
			return
		default:
		}

		var reqJ logoutReq
		if err := common.ReadJSON(r, &reqJ); err != nil {
			common.WriteError(w, http.StatusBadRequest, "Invalid body")
			return
		}

		req := &proto.LogoutRequest{
			Tokens: &proto.Tokens{
				Access:  reqJ.Tokens.Access,
				Refresh: reqJ.Tokens.Refresh,
			},
		}

		_, err := authClient.Logout(ctx, req)
		if err != nil {
			common.WriteError(w, http.StatusInternalServerError, "")
			return
		}

		common.WriteJSON(w, http.StatusOK, "")
	}
}
