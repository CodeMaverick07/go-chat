package api

import (
	"encoding/json"
	"go-chat/internals/store"
	"go-chat/internals/utils"
	"log"
	"net/http"
	"time"
)

const AuthScope = "login"

type AuthHandler struct {
	Logger     *log.Logger
	UserStore  store.UserStore
	TokenStore store.TokenStore
	OTPStore   store.OTPstore
}
type loginPasswordReq struct {
	Value    string `json:"value"`
	Password string `json:"password"`
}
type loginOTPreq struct {
	Email string `json:"email"`
}
type verifyOTPreq struct {
	Email   string `json:"email"`
	OTP     string `json:"otp"`
	Purpose string `json:"purpose"`
}

func NewAuthHandler(logger *log.Logger, userStore store.UserStore, tokenStore store.TokenStore, OTPStore store.OTPstore) *AuthHandler {
	return &AuthHandler{
		Logger:     logger,
		UserStore:  userStore,
		TokenStore: tokenStore,
		OTPStore:   OTPStore,
	}
}

func (h *AuthHandler) LoginWithEmailOrUsernameAndPassword(w http.ResponseWriter, r *http.Request) {
	var req loginPasswordReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Value == "" {
		h.Logger.Printf("Error:error while decoding %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "no value provided"})
		return
	}
	user, err := h.UserStore.GetUserByUserNameOrEmail(req.Value)

	if err != nil || user == nil {
		h.Logger.Printf("Error:user not found %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "user not found"})
		return
	}
	err = utils.VerifyHash(user.Password, req.Password)
	if err != nil {
		h.Logger.Printf("Error:incorrect password %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "incorrect password"})
		return
	}
	token, err := h.TokenStore.CreateNewToken(user.ID, 24*time.Hour, user.Scope)
	if err != nil {
		h.Logger.Printf("Error:error while creating token %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "internal server error"})
		return
	}
	var payload = map[string]interface{}{
		"user_id":    user.ID,
		"auth_token": token,
	}

	utils.WriteJSON(w, http.StatusAccepted, payload)
}

func (h *AuthHandler) LoginWithEmailandOTP(w http.ResponseWriter, r *http.Request) {
	var req loginOTPreq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Email == "" {
		h.Logger.Printf("Error:error while decoding %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "no value provided"})
		return
	}

	user, err := h.UserStore.GetUserByUserNameOrEmail(req.Email)
	if err != nil || user == nil {
		h.Logger.Printf("Error:user not found %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "user not found"})
		return
	}
	err = h.OTPStore.SendOTP(user.UserName, user.Email, AuthScope)
	if err != nil {
		h.Logger.Printf("Error:error while sending otp %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "internal server error"})
		return
	}
	utils.WriteJSON(w, http.StatusAccepted, utils.Envelope{
		"message": "OTP sent successfully",
	})
}

func (h *AuthHandler) VerifyLoginOTP(w http.ResponseWriter, r *http.Request) {
	var req verifyOTPreq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || len(req.OTP) != 6 {
		h.Logger.Printf("Error:otp not provided %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "no value provided"})
		return
	}
	if req.Email == "" {
		h.Logger.Printf("Error:email not provided %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "no value provided"})
		return
	}
	user, err := h.UserStore.GetUserByUserNameOrEmail(req.Email)
	if err != nil || user == nil {
		h.Logger.Printf("Error:user not found %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "user not found"})
		return
	}

	if req.Purpose != AuthScope {
		h.Logger.Printf("Error:wrong purpose  %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "no value provided"})
		return
	}
	_, err = h.OTPStore.VerifyOTP(req.Email, req.OTP, store.OTPPurpose(req.Purpose))
	if err != nil {
		h.Logger.Printf("Error:wrong otp  %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "wrong otp"})
		return
	}
	token, err := h.TokenStore.CreateNewToken(user.ID, 24*time.Hour, user.Scope)
	if err != nil {
		h.Logger.Printf("Error:error while creating token %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "internal server error"})
		return
	}
	var payload = map[string]interface{}{
		"user_id":    user.ID,
		"auth_token": token,
	}

	utils.WriteJSON(w, http.StatusAccepted, payload)

}
