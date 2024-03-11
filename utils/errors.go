package utils

import (
	"net/http"
)

type LibError struct {
	Error string `json:"error"`
}

func PermissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, LibError{Error: "Permission denied"})
}

func MethodNotAllowed(w http.ResponseWriter) error {
	return WriteJSON(w, http.StatusForbidden, LibError{Error: "Method not allowed"})
}

func NotAuthenticated(w http.ResponseWriter) error {
	return WriteJSON(w, http.StatusUnauthorized, LibError{Error: "Not authenticated"})
}
