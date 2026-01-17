package api

import (
	"encoding/json"
	"go-chat/internals/middleware"
	"go-chat/internals/store"
	"go-chat/internals/utils"
	"log"
	"net/http"
	"time"
	"unicode/utf8"
)

type VerifyUserAndRegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	OTP      string `json:"otp"`
	Purpose  string `json:"purpose"`
}
type SendOTPRequest struct {
	Email   string `json:"email"`
	Purpose string `json:"purpose"`
}

type UserHandler struct {
	UserStore  store.UserStore
	Logger     *log.Logger
	OTPStore   store.OTPstore
	TokenStore store.TokenStore
}

func NewUserHandler(userStore store.UserStore, logger *log.Logger, authStore store.OTPstore, tokenStore store.TokenStore) *UserHandler {
	return &UserHandler{
		UserStore:  userStore,
		Logger:     logger,
		OTPStore:   authStore,
		TokenStore: tokenStore,
	}
}

func (u *UserHandler) SendOTPHandler(w http.ResponseWriter, r *http.Request) {
	var req SendOTPRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		u.Logger.Println("not able to decode send otp", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": err.Error()})
		return
	}
	if req.Email == "" {
		utils.WriteJSON(
			w,
			http.StatusBadRequest,
			utils.Envelope{"error": "email is required"},
		)
		return
	}
	if req.Purpose != "verify" && req.Purpose != "login" {
		utils.WriteJSON(
			w,
			http.StatusBadRequest,
			utils.Envelope{"error": "invalid otp purpose"},
		)
		return
	}
	err = u.UserStore.IsUniqueUsernameOrEmail(req.Email, "email")
	if err != nil {
		u.Logger.Println("email is not unique", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": err.Error()})
		return
	}

	err = u.OTPStore.SendOTP("user", req.Email, store.OTPPurpose(req.Purpose))
	if err != nil {
		u.Logger.Println("not able to send otp", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": err.Error()})
		return
	}
	utils.WriteJSON(w, 200, utils.Envelope{"data": req.Email})
}

func (u *UserHandler) VerifyOTPHandler(w http.ResponseWriter, r *http.Request) {

}

func (u *UserHandler) VerifyOTPAndCreateUserHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req VerifyUserAndRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		u.Logger.Println("decode request:", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"error": "invalid request body",
		})
		return
	}
	if req.Email == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"error": "email is required",
		})
		return
	}

	if utf8.RuneCountInString(req.Password) < 8 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"error": "password must be at least 8 characters",
		})
		return
	}

	if req.Username == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"error": "username is required",
		})
		return
	}

	if err := u.UserStore.IsUniqueUsernameOrEmail(req.Username, "username"); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"error": "username already taken",
		})
		return
	}

	if req.Purpose != "verify" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"error": "invalid otp purpose",
		})
		return
	}

	if len(req.OTP) != 6 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"error": "invalid otp",
		})
		return
	}
	if _, err := u.OTPStore.VerifyOTP(
		req.Email,
		req.OTP,
		store.OTPPurpose(req.Purpose),
	); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"error": "invalid or expired otp",
		})
		return
	}
	passwordHash, err := utils.Hash(req.Password)
	if err != nil {
		u.Logger.Println("hash password:", err)
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}
	user := store.User{
		UserName:  req.Username,
		Email:     req.Email,
		Password:  passwordHash,
		Scope:     utils.UserScope,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := u.UserStore.CreateUser(&user); err != nil {
		u.Logger.Printf("Error: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"error": "internal server error",
		})
		return
	}
	token, err := u.TokenStore.CreateNewToken(user.ID, 24*time.Hour, user.Scope)
	if err != nil {
		u.Logger.Printf("Error: %v", err)
		utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{
			"error": "internal server error",
		})
		return
	}
	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{
		"auth_token": token,
	})
}
func (u *UserHandler) WebsocketTokenHandler(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user.IsAnonymousUser() {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "user is not logged in"})
		return
	}
	token, err := u.TokenStore.CreateNewToken(user.ID, time.Hour*10, utils.SocketScope)
	if err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": err.Error()})
		return
	}
	utils.WriteJSON(w, http.StatusAccepted, utils.Envelope{"socket_token": token})
}
