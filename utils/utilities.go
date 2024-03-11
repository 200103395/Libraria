package utils

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"math/rand"
	"net/http"
	"strconv"
)

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func MakeToken() string {
	var alphaNumRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	emailVerRandRunes := make([]rune, 64)
	for i := 0; i < 64; i++ {
		emailVerRandRunes[i] = alphaNumRunes[rand.Intn(len(alphaNumRunes)-1)]
	}
	return string(emailVerRandRunes)
}

func GetID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idStr)
	}
	return id, nil
}

func GetMAIL(r *http.Request) string {
	mail := mux.Vars(r)["email"]
	return mail
}

func GetTAG(r *http.Request) string {
	tag := mux.Vars(r)["tag"]
	return tag
}
