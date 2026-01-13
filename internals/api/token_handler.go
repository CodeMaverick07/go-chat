package api

import (
	"encoding/json"
	"go-chat/internals/store"
	"go-chat/internals/tokens"
	"go-chat/internals/utils"
	"log"
	"net/http"
	"time"
)

type createTokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenHandler struct {
	TokenStore store.TokenStore
	UserStore  store.UserStore
	Logger     *log.Logger
}

func NewTokenHandler(tokenStore store.TokenStore, userStore store.UserStore, logger *log.Logger) *TokenHandler {
	return &TokenHandler{
		TokenStore: tokenStore,
		UserStore:  userStore,
		Logger:     logger,
	}
}

func (h *TokenHandler) HandleCreateToken(w http.ResponseWriter, r *http.Request) {
	var req createTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.Logger.Println("ERROR:error while creating token", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "internal server error"})
		return
	}
	user, err := h.UserStore.GetUserByUserNameOrEmail(req.Username)
	if err != nil || user == nil {
		h.Logger.Println("ERROR:no user found", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "internal server error"})
		return
	}

	err = utils.VerifyHash(user.Password, req.Password)
	if err != nil {
		h.Logger.Println("ERROR:password dose not match", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "internal server error"})
		return
	}

	token, err := h.TokenStore.CreateNewToken(user.ID, 24*time.Hour, tokens.ScopeAuth)
	if err != nil {
		h.Logger.Println("ERROR:not able to create token", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "internal server error"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"auth_token": token})
}
