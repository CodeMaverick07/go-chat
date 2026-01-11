package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Envelope map[string]interface{}

func WriteJSON(w http.ResponseWriter,status int, data Envelope) error {
	js, err := json.MarshalIndent(data,""," ")
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