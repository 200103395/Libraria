package main

import (
	"encoding/json"
	"errors"
	"fmt"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)

type LibServer struct {
	listenAddr string
	store      Storage
	email      Email
}

func NewLibServer(listenAddr string, store Storage, email Email) *LibServer {
	return &LibServer{
		listenAddr: listenAddr,
		store:      store,
		email:      email,
	}
}

func (s *LibServer) Run() {
	r := mux.NewRouter()
	r.HandleFunc("/", makeHTTPHandleFunc(s.HomeHandler))
	r.HandleFunc("/about", makeHTTPHandleFunc(s.HomeHandler))
	r.HandleFunc("/search", makeHTTPHandleFunc(s.HomeHandler)) // search & filter

	r.HandleFunc("/account/confirm/{email}/{tag}", makeHTTPHandleFunc(s.AccountConfirm))

	r.HandleFunc("/account/register", makeHTTPHandleFunc(s.AccountCreateHandler)) // Email confirmation + new db table
	r.HandleFunc("/account/login", makeHTTPHandleFunc(s.AccountLoginHandler))     // JWT storing information
	r.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandleFunc(s.AccountHandler), s.store))
	r.HandleFunc("/account", makeHTTPHandleFunc(s.AccountsHandler))             // Why?
	r.HandleFunc("/password_reset", makeHTTPHandleFunc(s.PasswordResetHandler)) // Email sent + new query

	r.HandleFunc("/library/register", makeHTTPHandleFunc(s.LibraryCreateHandler)) // Email conf + new db table
	r.HandleFunc("/library/login", makeHTTPHandleFunc(s.LibraryLoginHandler))     // JWT storing information
	r.HandleFunc("/library/{id}", makeHTTPHandleFunc(s.GetLibraryHandler))        // map API
	r.HandleFunc("/library", makeHTTPHandleFunc(s.LibraryHandler))

	r.HandleFunc("/book/{id}", makeHTTPHandleFunc(s.HomeHandler)) // book get
	r.HandleFunc("/book", makeHTTPHandleFunc(s.HomeHandler))

	if err := http.ListenAndServe(s.listenAddr, r); err != nil {
		log.Fatal(err)
	}
}

func (c *LibServer) PasswordResetHandler(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *LibServer) HomeHandler(w http.ResponseWriter, r *http.Request) error {
	w.Write([]byte("Hello, World!"))
	return nil
}

func (s *LibServer) LibraryHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		//global
	}
	if r.Method == "POST" {
		//filter
	}
	return nil
}

func (s *LibServer) LibraryCreateHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return errors.New("Method not allowed")
	}

	var library LibraryAccount
	err := json.NewDecoder(r.Body).Decode(&library)
	w.Header().Add("Content-Type", "application/json")
	if err != nil {
		return err
	}
	isExists, err := s.store.CheckEmail(library.Email)
	if err != nil {
		return err
	}
	if isExists == false {
		WriteJSON(w, http.StatusBadRequest, fmt.Sprintf("Email %s is already in use", library.Email))
		return nil
	}
	err = NewLibraryAccount(&library)
	err = s.store.CreateLibraryAccount(&library)
	if err != nil {
		return err
	}
	json.NewEncoder(w).Encode(library)
	return nil
}

func (s *LibServer) AccountHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.GetAccountHandler(w, r)
	}
	return errors.New("Method not allowed")
}

func (s *LibServer) GetAccountHandler(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}
	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}
	ans := []any{"User is logged in", account}
	WriteJSON(w, http.StatusOK, ans)
	/*data, _ := json.Marshal(accounts)
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)*/
	return nil
}

func (s *LibServer) GetLibraryHandler(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}
	library, err := s.store.GetLibraryByID(id)
	if err != nil {
		return err
	}
	data, _ := json.Marshal(library)
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
	return nil
}

func (s *LibServer) AccountsHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.GetAccountsHandler(w, r)
	}
	return errors.New("Method not allowed")
}

func (s *LibServer) GetAccountsHandler(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}
	data, _ := json.Marshal(accounts)
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
	return nil
}

func (s *LibServer) AccountLoginHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	acc, err := s.store.GetAccountByEmail(req.Email)
	if err != nil {
		return err
	}

	if !acc.ValidPassword(req.Password) {
		return fmt.Errorf("not authenticated")
	}

	token, err := createJWT(acc)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, token)
}

func (s *LibServer) LibraryLoginHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	acc, err := s.store.GetLibraryByEmail(req.Email)
	if err != nil {
		return err
	}

	if !acc.ValidPassword(req.Password) {
		return fmt.Errorf("not authenticated")
	}

	token, err := createLibraryJWT(acc)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, token)
}

func (s *LibServer) AccountCreateHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return errors.New("Method not allowed")
	}

	var account Account
	err := json.NewDecoder(r.Body).Decode(&account)
	w.Header().Add("Content-Type", "application/json")
	if err != nil {
		return err
	}
	isExists, err := s.store.CheckEmail(account.Email)
	if err != nil {
		return err
	}
	if isExists == false {
		WriteJSON(w, http.StatusBadRequest, fmt.Sprintf("Email %s is already in use", account.Email))
		return nil
	}

	var alphaNumRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	emailVerRandRunes := make([]rune, 64)
	for i := 0; i < 64; i++ {
		emailVerRandRunes[i] = alphaNumRunes[rand.Intn(len(alphaNumRunes)-1)]
	}
	print(emailVerRandRunes)

	req := CreateRequest{
		FirstName:     account.FirstName,
		LastName:      account.LastName,
		Password:      account.Password,
		Email:         account.Email,
		Address:       account.Address,
		ContactNumber: account.ContactNumber,
		Tag:           string(emailVerRandRunes),
	}
	err = s.store.CreateRequest(&req)
	if err != nil {
		return err
	}
	s.email.EmailConfirmationMessage([]string{req.Email}, req.FirstName, req.LastName, "localhost:8000/account/confirm/"+req.Email+"/"+req.Tag)
	ans := []any{fmt.Sprintf("Message was sent to %s , to confirm account please follow the instructions in the message", req.Email), req}
	WriteJSON(w, http.StatusOK, ans)
	return nil
}

func (s *LibServer) AccountConfirm(w http.ResponseWriter, r *http.Request) error {
	mail := getMAIL(r)
	tag := getTAG(r)
	user, err := s.store.GetRequestByMAIL(mail)
	if err != nil {
		return err
	}
	if user.Tag != tag {
		WriteJSON(w, http.StatusBadRequest, "Incorrect url") // Change url
	}
	account := &Account{
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Email:         user.Email,
		Password:      user.Password,
		Address:       user.Address,
		ContactNumber: user.ContactNumber,
	}
	err = NewAccount(account)
	err = s.store.CreateAccount(account)
	if err != nil {
		return err
	}
	s.store.DeleteRequest(user)
	ans := []any{fmt.Sprintf("User %s have been successfully confirmed", account.FirstName), account}
	WriteJSON(w, http.StatusOK, ans)
	return nil
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

func makeHTTPHandleFunc(f LibFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, LibError{Error: err.Error()})
		}
	}
}

func getID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idStr)
	}
	return id, nil
}

func getMAIL(r *http.Request) string {
	mail := mux.Vars(r)["email"]
	return mail
}

func getTAG(r *http.Request) string {
	tag := mux.Vars(r)["tag"]
	return tag
}

// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50SUQiOjEsImV4cGlyZXNBdCI6MTUwMDB9.xkJ1gVSN-xQK5xVmmsTN_rKMoLLm0xwz0lqjzYu5WoI
func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling JWT auth middleware")

		tokenString := r.Header.Get("x-jwt-token")
		token, err := validateJWT(tokenString)
		if err != nil {
			permissionDenied(w)
			return
		}
		if !token.Valid {
			permissionDenied(w)
			return
		}
		userID, err := getID(r)
		if err != nil {
			permissionDenied(w)
			return
		}
		account, err := s.GetAccountByID(userID)
		if err != nil {
			permissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if account.ID != uint(claims["accountID"].(float64)) {
			permissionDenied(w)
			return
		}

		if err != nil {
			WriteJSON(w, http.StatusForbidden, LibError{Error: "invalid token"})
			return
		}

		handlerFunc(w, r)
	}
}

var ErrorUnauthorized = errors.New("Unauthorized")

func readJWT(r *http.Request, s Storage) (*Account, error) {
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
	if claims["accountID"] == nil {
		return nil, ErrorUnauthorized
	}

	account, err := s.GetAccountByID(int(claims["accountID"].(float64)))
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
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(secret), nil
	})
}

func createJWT(account *Account) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt": 15000,
		"accountID": account.ID,
	}

	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func createLibraryJWT(account *LibraryAccount) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt": 15000,
		"accountID": account.ID,
	}

	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}
