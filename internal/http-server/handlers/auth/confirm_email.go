package auth

import (
	"net/http"

	proto "github.com/GP-Hacks/proto/pkg/api/auth"
	"github.com/go-chi/chi/v5"
)

func NewConfirmEmailPageHandler(authClient proto.AuthServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		select {
		case <-ctx.Done():
			w.WriteHeader(http.StatusRequestTimeout)
			w.Write([]byte("<html><body><h1>Request timed out</h1></body></html>"))
			return
		default:
		}

		token := chi.URLParam(r, "token")
		if token == "" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`
   			<html>
			<head><meta charset="utf-8"/></head>
   			<body>
   				<h1>Ошибка подтверждения email</h1>
   				<p>Неверная ссылка для подтверждения</p>
   			</body>
   			</html>
   		`))
			return
		}

		req := &proto.ConfirmEmailRequest{
			Token: token,
		}

		_, err := authClient.ConfirmEmail(ctx, req)

		w.Header().Set("Content-Type", "text/html")

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`
   			<html>
						<head><meta charset="utf-8"/></head>
   			<body>
   				<h1>Ошибка подтверждения email</h1>
   				<p>Неверный или истекший токен подтверждения</p>
   			</body>
   			</html>
   		`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
   		<html>
					<head><meta charset="utf-8"/></head>
   		<body>
   			<h1>Email успешно подтвержден!</h1>
   			<p>Ваш email адрес был успешно подтвержден. Теперь вы можете войти в систему.</p>
   		</body>
   		</html>
   	`))
	}
}
