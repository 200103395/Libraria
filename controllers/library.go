package controllers

import (
	"Libraria/types"
	"Libraria/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (s *LibServer) LibraryHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	libraries, err := s.store.GetLibraries()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, libraries)
	// return HTML file
}

func (s *LibServer) LibraryCreateHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" && r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	if r.Method == "GET" {
		// return HTML file
	}
	var library types.LibraryAccount
	err := json.NewDecoder(r.Body).Decode(&library)
	if err != nil {
		return err
	}
	err = library.ValidateAccount()
	if err != nil {
		return err
	}
	isExists, err := s.store.CheckEmail(library.Email)
	if err != nil {
		return err
	}
	if isExists == false {
		return WriteJSON(w, http.StatusBadRequest, fmt.Sprintf("Email %s is already in use", library.Email))
	}

	expiresAt := time.Now().Add(expiration * time.Minute).UTC()
	req := types.LibRequest{
		Name:          library.Name,
		Password:      library.Password,
		Email:         library.Email,
		Address:       library.Address,
		ContactNumber: library.ContactNumber,
		Tag:           utils.MakeToken(),
		ExpiresAt:     expiresAt,
	}
	err = s.store.CreateLibRequest(&req)
	if err != nil {
		return err
	}
	appeal := req.Name + " Library"
	err = s.email.EmailConfirmationMessage(req.Email, appeal, domain+"/account/confirm/"+req.Tag)
	if err != nil {
		return err
	}
	ans := []any{fmt.Sprintf("Message was sent to %s , to confirm account please follow the instructions in the message", req.Email), req}
	return WriteJSON(w, http.StatusOK, ans)
	// show success HTML page

}

func (s *LibServer) LibraryConfirmHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	tag := utils.GetTAG(r)
	lib, err := s.store.GetLibRequestByTAG(tag)
	if err != nil {
		return err
	}
	library := &types.LibraryAccount{
		Name:          lib.Name,
		Email:         lib.Email,
		Password:      lib.Password,
		Address:       lib.Address,
		ContactNumber: lib.ContactNumber,
	}
	err = library.PasswordHash()
	if err != nil {
		return err
	}
	err = s.store.CreateLibraryAccount(library)
	if err != nil {
		return err
	}
	err = s.store.DeleteLibRequest(lib)
	if err != nil {
		// DO SOMETHING
	}
	ans := []any{fmt.Sprintf("Library %s have been successfully confirmed", library.Name), library}
	return WriteJSON(w, http.StatusOK, ans)
	// show success page
}

func (s *LibServer) GetLibraryHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	id, err := utils.GetID(r)
	if err != nil {
		return err
	}
	library, err := s.store.GetLibraryByID(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, library)
	// show library page HTML
}

func (s *LibServer) LibraryLoginHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" && r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	if r.Method == "GET" {
		// return HTML file
	}

	var req types.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	lib, err := s.store.GetLibraryByEmail(req.Email)
	if err != nil {
		return err
	}

	if !lib.ValidPassword(req.Password) {
		return fmt.Errorf("not authenticated")
	}

	token, err := createJWT(lib.Email)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, token)
	// show success page
}
