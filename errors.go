package main

import "net/http"

func permissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, LibError{Error: "permission denied"})
}
