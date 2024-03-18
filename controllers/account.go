package controllers

import (
	"Libraria/types"
	"Libraria/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	html, err := os.ReadFile("static/accountGet.html")
	if err != nil {
		return err
	}
	jsonAcc := struct {
		FirstName     string `json:"firstName"`
		LastName      string `json:"lastName"`
		Email         string `json:"email"`
		Address       string `json:"address"`
		ContactNumber string `json:"contactNumber"`
	}{
		account.FirstName,
		account.LastName,
		account.Email,
		account.Address,
		account.ContactNumber,
	}
	jsonAccMarsh, err := json.Marshal(jsonAcc)
	if err != nil {
		return err
	}
	if _, err = fmt.Fprintf(w, "<script id=\"headScript\">var account = %s;</script>", jsonAccMarsh); err != nil {
		return err
	}
	if _, err = fmt.Fprintf(w, string(html)); err != nil {
		return err
	}
	return nil
	// show account HTML page
}

func (s *LibServer) AccountLoginHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" && r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	if r.Method == "GET" {
		html, err := os.ReadFile("static/accountLogin.html")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, string(html))
		return err
		//return WriteJSON(w, http.StatusOK, string(html))
		// return HTML file
	}
	var req types.LoginRequest
	fmt.Println(req)
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
	cookie := http.Cookie{
		Name:     "x-jwt-token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(time.Minute * 15),
	}
	http.SetCookie(w, &cookie)
	return WriteJSON(w, http.StatusOK, cookie)
	// redirect
}

func (s *LibServer) AccountCreateHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" && r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	if r.Method == "GET" {
		html, err := os.ReadFile("static/accountRegister.html")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, string(html))
		return err
	}

	var account types.Account
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		return err
	}
	if err := account.ValidateAccount(); err != nil {
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
		Tag:           s.store.MakeToken("user_requests"),
		ExpiresAt:     expiresAt,
	}
	fmt.Println(req)
	err = s.store.CreateUserRequest(&req)
	if err != nil {
		return err
	}
	appeal := req.FirstName + " " + req.LastName
	err = s.email.EmailConfirmationMessage(req.Email, appeal, domain+"/account/confirm/"+req.Tag)
	fmt.Println(err)
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
	html, err := os.ReadFile("static/accountConfirm.html")
	_, err = fmt.Fprintf(w, string(html))
	return err
}
