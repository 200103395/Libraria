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

func (s *LibServer) PasswordResetConfirmHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" && r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	if r.Method == "GET" {
		html, err := os.ReadFile("static/passwordConfirm.html")
		if err != nil {
			return err
		}
		token := utils.GetTAG(r)
		_, err = fmt.Fprintf(w, fmt.Sprintf("<script> var token = \"%s\"; </script>", token))
		_, err = fmt.Fprintf(w, string(html))
		return err
	}
	token := utils.GetTAG(r)
	req, err := s.store.GetPasswordReset(token)
	fmt.Println(req, err, token)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, "Incorrect link") //relocate
	}
	var newPw types.NewPassword
	if err = json.NewDecoder(r.Body).Decode(&newPw); err != nil {
		return WriteJSON(w, http.StatusBadRequest, "Incorrect link")
	}
	fmt.Println(newPw)
	if newPw.NewPassword != newPw.NewPasswordConfirm {
		return WriteJSON(w, http.StatusBadRequest, "Incorrect link")
	}

	user, err1 := s.store.GetAccountByEmail(req.Email)
	library, err2 := s.store.GetLibraryByEmail(req.Email)
	fmt.Println(err1, err2)
	if err1 != nil && err2 != nil {
		return WriteJSON(w, http.StatusBadRequest, "Incorrect link") //relocate
	}
	if err1 == nil {
		user.Password = newPw.NewPassword
		err = user.PasswordHash()
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, "Incorrect link") //relocate
		}
		s.store.UpdateAccount(user)
	} else {
		library.Password = newPw.NewPassword
		err = library.PasswordHash()
		if err != nil {
			return WriteJSON(w, http.StatusBadRequest, "Incorrect link") //relocate
		}
		s.store.UpdateLibrary(library)
	}
	return WriteJSON(w, http.StatusOK, "Password has been changed")
}

func (s *LibServer) PasswordResetHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" && r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	if r.Method == "GET" {
		html, err := os.ReadFile("static/passwordReset.html")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, string(html))
		return err
	}

	var jspost struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&jspost); err != nil {
		return err
	}
	if err := s.store.CheckForRequest(jspost.Email); err != nil {
		return err
	}
	user, err1 := s.store.GetAccountByEmail(jspost.Email)
	library, err2 := s.store.GetLibraryByEmail(jspost.Email)
	if err1 != nil && err2 != nil {
		return WriteJSON(w, http.StatusBadRequest, "Incorrect email")
	}

	expiresAt := time.Now().Add(expiration * time.Minute).UTC()
	request := &types.PasswordResetRequest{
		Email:     jspost.Email,
		Token:     s.store.MakeToken("password_reset"),
		ExpiresAt: expiresAt,
	}
	appeal := ""
	if err1 != nil {
		appeal = library.Name + " Library"
	}
	if err2 != nil {
		appeal = user.FirstName + " " + user.LastName
	}
	err := s.email.PasswordResetMessage(request.Email, appeal, domain+"/password_reset/"+request.Token)

	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, "Error sending email, please try again later")
	}
	err = s.store.CreatePasswordReset(request)

	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, "Error connecting to db, please try again later")
	}
	return WriteJSON(w, http.StatusOK, "Message has been sent to email: "+request.Email)
	// show HTML success page
}
