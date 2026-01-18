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
		token := r.URL.Query().Get("token")
		if token == "" {
			utils.WriteJSON(w, http.StatusUnauthorized,
				utils.Envelope{"error": "invalid userId"})
			return
		}
		user, err := wm.UserStore.GetUserToken(utils.SocketScope, token)
		if err != nil {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "user not present"})
			return
		}
		ctx := context.WithValue(r.Context(), contexkeys.UserID, user.ID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})

}
