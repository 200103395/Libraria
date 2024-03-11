package controllers

import (
	"Libraria/types"
	"Libraria/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (s *LibServer) AccountHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.GetAccountHandler(w, r)
	}
	return utils.MethodNotAllowed(w)
}

func (s *LibServer) GetAccountHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	id, err := utils.GetID(r)
	if err != nil {
		return err
	}
	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}
	ans := []any{"User is logged in", account}
	return WriteJSON(w, http.StatusOK, ans)
	// show account HTML page
}

func (s *LibServer) AccountLoginHandler(w http.ResponseWriter, r *http.Request) error {
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

	acc, err := s.store.GetAccountByEmail(req.Email)
	if err != nil {
		return err
	}

	if !acc.ValidPassword(req.Password) {
		return utils.NotAuthenticated(w)
	}

	token, err := createJWT(acc.Email)
	if err != nil {
		return err
	}

	ans := []any{"User have been successfully logged in", token}

	return WriteJSON(w, http.StatusOK, ans)
	// redirect to main page
}

func (s *LibServer) AccountCreateHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" && r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	if r.Method == "GET" {
		// return HTML file
	}

	var account types.Account
	err := json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		return err
	}
	err = account.ValidateAccount()
	if err != nil {
		return err
	}
	isExists, err := s.store.CheckEmail(account.Email)
	if err != nil {
		return err
	}
	if isExists == false {
		return WriteJSON(w, http.StatusBadRequest, fmt.Sprintf("Email %s is already in use", account.Email))
	}

	expiresAt := time.Now().Add(expiration * time.Minute).UTC()
	req := types.UserRequest{
		FirstName:     account.FirstName,
		LastName:      account.LastName,
		Password:      account.Password,
		Email:         account.Email,
		Address:       account.Address,
		ContactNumber: account.ContactNumber,
		Tag:           utils.MakeToken(),
		ExpiresAt:     expiresAt,
	}
	err = s.store.CreateUserRequest(&req)
	if err != nil {
		return err
	}
	appeal := req.FirstName + " " + req.LastName
	err = s.email.EmailConfirmationMessage(req.Email, appeal, domain+"/account/confirm/"+req.Tag)
	if err != nil {
		return err
	}
	ans := []any{fmt.Sprintf("Message was sent to %s , to confirm account please follow the instructions in the message", req.Email), req}
	return WriteJSON(w, http.StatusOK, ans)
	// show success HTML page
}

func (s *LibServer) AccountConfirm(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	tag := utils.GetTAG(r)
	user, err := s.store.GetUserRequestByTAG(tag)
	if err != nil {
		return err
	}
	account := &types.Account{
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Email:         user.Email,
		Password:      user.Password,
		Address:       user.Address,
		ContactNumber: user.ContactNumber,
	}
	err = account.PasswordHash()
	if err != nil {
		return err
	}
	err = s.store.CreateAccount(account)
	if err != nil {
		return err
	}
	err = s.store.DeleteUserRequest(user)
	if err != nil {
		// do something
	}
	ans := []any{fmt.Sprintf("User %s have been successfully confirmed", account.FirstName), account}
	return WriteJSON(w, http.StatusOK, ans)
	// return HTML success page
}
