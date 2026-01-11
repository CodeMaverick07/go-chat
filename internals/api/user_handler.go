package api

import (
	"encoding/json"
	"go-chat/internals/store"
	"go-chat/internals/utils"
	"log"
	"net/http"
	"time"
)

type RegisterUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserHandler struct {
	UserStore store.UserStore
	Logger *log.Logger
}

func NewUserHandler(UserStore store.UserStore, Logger *log.Logger) *UserHandler {
	return &UserHandler{
		UserStore: UserStore,
		Logger: Logger,
	}
}

func (u *UserHandler) CreateUserHandler(w http.ResponseWriter,r *http.Request){
	var req RegisterUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		u.Logger.Fatal("not able to create user 1",err)
		utils.WriteJSON(w,http.StatusBadRequest,utils.Envelope{"error":err.Error()})
		return
	}
	user := store.User{
		UserName: req.Username,
		Email: req.Email,
		Password: req.Password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = u.UserStore.CreateUser(&user)
	if err != nil {
		u.Logger.Fatal("not able to create user 2",err)
		utils.WriteJSON(w,http.StatusBadRequest,utils.Envelope{"error":err.Error()})
		return
	}
	utils.WriteJSON(w, http.StatusCreated, utils.Envelope{"data": user})

}