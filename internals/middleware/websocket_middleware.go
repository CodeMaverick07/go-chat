package middleware

import (
	"fmt"
	"go-chat/internals/store"
	"go-chat/internals/utils"
	"net/http"
)

type WebsocketMiddleware struct {
	UserStore store.UserStore
}

func (wm *WebsocketMiddleware) AuthenticateWebsockets(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("1")
		token := r.URL.Query().Get("token")
		fmt.Println("2", token)
		if token == "" {
			utils.WriteJSON(w, http.StatusUnauthorized,
				utils.Envelope{"error": "invalid userId"})
			return
		}
		user, err := wm.UserStore.GetUserToken(utils.SocketScope, token)
		if err != nil {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "user not present"})
		}
		fmt.Print(user)
		next.ServeHTTP(w, r)
	})

}
