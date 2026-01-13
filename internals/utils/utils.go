package utils

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

type Envelope map[string]interface{}

func WriteJSON(w http.ResponseWriter, status int, data Envelope) error {
	js, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}
	js = append(js, '\n')
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func ReadParamID(r *http.Request) (int64, error) {
	paramId := chi.URLParam(r, "id")
	if paramId == "" {
		return 0, errors.New("invalid id parameter")
	}
	id, err := strconv.ParseInt(paramId, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

type OTP struct {
	Code      string
	ExpiresAt time.Time
}

func GenerateOTP() (*OTP, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return nil, err
	}

	code := fmt.Sprintf("%06d", n.Int64())

	return &OTP{
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}, nil
}

func Hash(value string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyHash(hash, value string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(value),
	)
}
