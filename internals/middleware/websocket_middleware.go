package middleware

import (
	"context"
	"go-chat/internals/contexkeys"
	"go-chat/internals/store"
	"go-chat/internals/utils"
	"net/http"
)

type WebsocketMiddleware struct {
	UserStore store.UserStore
}

func (wm *WebsocketMiddleware) AuthenticateWebsockets(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println("1")
		token := r.URL.Query().Get("token")
		if token == "" {
			utils.WriteJSON(w, http.StatusUnauthorized,
				utils.Envelope{"error": "invalid userId"})
			return
		}
		//fmt.Println("2", token)

		user, err := wm.UserStore.GetUserToken(utils.SocketScope, token)
		if err != nil || user == nil {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "user not present"})
			return
		}
		//fmt.Println("3", user)

		ctx := context.WithValue(r.Context(), contexkeys.UserID, user.ID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
