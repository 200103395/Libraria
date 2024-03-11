package controllers

import (
	"Libraria/database"
	"Libraria/mail"
	"Libraria/types"
	"Libraria/utils"
	"encoding/json"
	"errors"
	"fmt"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	expiration = 20
	domain     = "http://localhost:8000"
)

var (
	ErrorUnauthorized = errors.New("unauthorized")
)

type LibServer struct {
	listenAddr string
	store      database.Storage
	email      mail.Email
}

func MakeHTTPHandleFunc(f LibFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, LibError{Error: err.Error()})
		}
	}
}

func NewLibServer(listenAddr string, store database.Storage, email mail.Email) *LibServer {
	return &LibServer{
		listenAddr: listenAddr,
		store:      store,
		email:      email,
	}
}

func (s *LibServer) Run() {
	r := mux.NewRouter()
	r.HandleFunc("/", MakeHTTPHandleFunc(s.HomeHandler)) // get books & libs
	r.HandleFunc("/about", MakeHTTPHandleFunc(s.AboutHandler))
	r.HandleFunc("/search", MakeHTTPHandleFunc(s.HomeHandler)) // search & filter

	r.HandleFunc("/account/confirm/{tag}", MakeHTTPHandleFunc(s.AccountConfirm))
	r.HandleFunc("/account/register", MakeHTTPHandleFunc(s.AccountCreateHandler))
	r.HandleFunc("/account/login", MakeHTTPHandleFunc(s.AccountLoginHandler))
	r.HandleFunc("/account/{id}", withJWTAuth(MakeHTTPHandleFunc(s.AccountHandler), s.store))

	r.HandleFunc("/password_reset/{tag}", MakeHTTPHandleFunc(s.PasswordResetConfirmHandler))
	r.HandleFunc("/password_reset", MakeHTTPHandleFunc(s.PasswordResetHandler))

	r.HandleFunc("/library/register", MakeHTTPHandleFunc(s.LibraryCreateHandler))
	r.HandleFunc("/library/login", MakeHTTPHandleFunc(s.LibraryLoginHandler))
	r.HandleFunc("/library/{id}", MakeHTTPHandleFunc(s.GetLibraryHandler))
	r.HandleFunc("/library", MakeHTTPHandleFunc(s.LibraryHandler))

	r.HandleFunc("/book/{id}", MakeHTTPHandleFunc(s.BookHandler))

	if err := http.ListenAndServe(s.listenAddr, r); err != nil {
		log.Fatal(err)
	}
}

func (s *LibServer) HomeHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return utils.MethodNotAllowed(w)
	}
	readJWT(r, s.store)
	_, err := w.Write([]byte("Hello, World!")) //return HTML for page
	return err
}

func (s *LibServer) AboutHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return utils.MethodNotAllowed(w)
	}
	_, err := w.Write([]byte("Hello, World!")) // return HTML for page
	return err
}

func (s *LibServer) PasswordResetConfirmHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" && r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	if r.Method == "GET" {
		// return HTML file
	}
	token := utils.GetTAG(r)
	req, err := s.store.GetPasswordReset(token)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, "Incorrect link") //relocate
	}
	var newPw types.NewPassword
	if err = json.NewDecoder(r.Body).Decode(&newPw); err != nil {
		return WriteJSON(w, http.StatusBadRequest, "Incorrect link")
	}
	if newPw.NewPassword != newPw.NewPasswordConfirm {
		return WriteJSON(w, http.StatusBadRequest, "Incorrect link")
	}

	user, err1 := s.store.GetAccountByEmail(req.Email)
	library, err2 := s.store.GetLibraryByEmail(req.Email)
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
		// return HTML file
	}

	var jspost struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&jspost); err != nil {
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
		Token:     utils.MakeToken(),
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

func (s *LibServer) BookHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	id, err := utils.GetID(r)
	if err != nil {
		return err
	}
	book, err := s.store.GetBookByID(id)
	return WriteJSON(w, http.StatusOK, book)
	// HTML show
}

type LibFunc func(w http.ResponseWriter, r *http.Request) error

type LibError struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func withJWTAuth(handlerFunc http.HandlerFunc, s database.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling JWT auth middleware")

		tokenString := r.Header.Get("x-jwt-token")
		token, err := validateJWT(tokenString)
		if err != nil {
			utils.PermissionDenied(w)
			return
		}
		if !token.Valid {
			utils.PermissionDenied(w)
			return
		}
		userID, err := utils.GetID(r)
		if err != nil {
			utils.PermissionDenied(w)
			return
		}
		account, err := s.GetAccountByID(userID)
		if err != nil {
			utils.PermissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if account.ID != uint(claims["accountID"].(float64)) {
			utils.PermissionDenied(w)
			return
		}

		if err != nil {
			WriteJSON(w, http.StatusForbidden, LibError{Error: "invalid token"})
			return
		}

		handlerFunc(w, r)
	}
}

func readJWT(r *http.Request, s database.Storage) (*types.Account, error) {
	// Maybe rewriting in middleware format
	tokenString := r.Header.Get("x-jwt-token")
	token, err := validateJWT(tokenString)
	if err != nil {
		return nil, ErrorUnauthorized
	}
	if !token.Valid {
		return nil, ErrorUnauthorized
	}
	claims := token.Claims.(jwt.MapClaims)
	if claims["email"] == nil {
		return nil, ErrorUnauthorized
	}
	fmt.Println(claims["email"].(string))
	account, err := s.GetAccountByEmail(claims["email"].(string))
	if err != nil {
		return nil, ErrorUnauthorized
	}

	return account, nil
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(secret), nil
	})
}

func createJWT(email string) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt": 15000,
		"email":     email,
	}

	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}
