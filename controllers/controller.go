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
	"io/ioutil"
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
	ErrorUnauthorized               = errors.New("unauthorized")
	loginTimeMinutes  time.Duration = 120
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
	r.HandleFunc("/search", MakeHTTPHandleFunc(s.SearchHandler)) // search & filter

	r.HandleFunc("/login", MakeHTTPHandleFunc(s.LoginHandler))
	r.HandleFunc("/register", MakeHTTPHandleFunc(s.RegisterHandler))
	r.HandleFunc("/account/settings", MakeHTTPHandleFunc(s.AccountSettingsHandler))
	r.HandleFunc("/account/confirm/{tag}", MakeHTTPHandleFunc(s.AccountConfirm))
	r.HandleFunc("/account/register", MakeHTTPHandleFunc(s.AccountCreateHandler))
	r.HandleFunc("/account/login", MakeHTTPHandleFunc(s.AccountLoginHandler))
	r.HandleFunc("/account/{id}", withJWTAuth(MakeHTTPHandleFunc(s.AccountHandler), s.store))

	r.HandleFunc("/password_reset/{tag}", MakeHTTPHandleFunc(s.PasswordResetConfirmHandler))
	r.HandleFunc("/password_reset", MakeHTTPHandleFunc(s.PasswordResetHandler))

	r.HandleFunc("/library/settings", MakeHTTPHandleFunc(s.LibrarySettingsHandler))
	r.HandleFunc("/library/confirm/{tag}", MakeHTTPHandleFunc(s.LibraryConfirmHandler))
	r.HandleFunc("/library/register", MakeHTTPHandleFunc(s.LibraryCreateHandler))
	r.HandleFunc("/library/login", MakeHTTPHandleFunc(s.LibraryLoginHandler))
	r.HandleFunc("/library/{id}", MakeHTTPHandleFunc(s.GetLibraryHandler))
	r.HandleFunc("/library", MakeHTTPHandleFunc(s.LibraryHandler))

	r.HandleFunc("/unAuthorize", MakeHTTPHandleFunc(s.UnAuthorizeHandler))
	r.HandleFunc("/getHeader", MakeHTTPHandleFunc(s.GetHeaderHandler))

	r.HandleFunc("/book/{id}", MakeHTTPHandleFunc(s.BookHandler))
	r.HandleFunc("/book/create", MakeHTTPHandleFunc(s.BookCreateHandler))

	r.HandleFunc("/getAuth", MakeHTTPHandleFunc(s.GetAuthHandler))
	r.HandleFunc("/getLibraries", MakeHTTPHandleFunc(s.GetAllLibrariesHandler))
	r.HandleFunc("/getSomeBooks", MakeHTTPHandleFunc(s.GetSomeBooksHandler))
	r.HandleFunc("/getLibrariesByBook/{id}", MakeHTTPHandleFunc(s.GetLibrariesByBookIDHandler))
	r.HandleFunc("/getLastBooks", MakeHTTPHandleFunc(s.GetLastBooks))
	r.HandleFunc("/getBooksByLibrary/{id}", MakeHTTPHandleFunc(s.GetBooksByLibraryIDHandler))

	if err := http.ListenAndServe(s.listenAddr, r); err != nil {
		log.Fatal(err)
	}
}

func (s *LibServer) HomeHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	index, err := ioutil.ReadFile("static/index.html")
	if err != nil {
		return err
	}
	// Convert JSON data to bytes
	if _, err = fmt.Fprintf(w, string(index)); err != nil {
		return err
	}
	_, err = w.Write([]byte("Hello, World!")) //return HTML for page
	return err
}

func (s *LibServer) AboutHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	html, err := os.ReadFile("static/about.html")
	if err != nil {
		return err
	}
	if _, err = fmt.Fprintf(w, string(html)); err != nil {
		return err
	}
	err = WriteJSON(w, http.StatusOK, "Hello, World! This is us")
	return err
}

func (s *LibServer) LoginHandler(w http.ResponseWriter, r *http.Request) error {
	html, err := os.ReadFile("static/login.html")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, string(html))
	return err
}

func (s *LibServer) RegisterHandler(w http.ResponseWriter, r *http.Request) error {
	html, err := os.ReadFile("static/register.html")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, string(html))
	return err
}

func (s *LibServer) UnAuthorizeHandler(w http.ResponseWriter, r *http.Request) error {
	deleteJWT(w)
	_, err := fmt.Fprintf(w, "<script>window.location.href = '"+domain+"'; </script>")
	return err
}

func (s *LibServer) GetHeaderHandler(w http.ResponseWriter, r *http.Request) error {
	html, err := os.ReadFile("static/header.html")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, string(html))
	return err
}

func (s *LibServer) GetAuthHandler(w http.ResponseWriter, r *http.Request) error {
	acc, err := readJWT(r, s.store)
	lib, err2 := readLibJWT(r, s.store)
	if err != nil && err2 != nil {
		return err
	}
	w.Header().Add("Content-Type", "application/json")
	if err == nil {
		return json.NewEncoder(w).Encode(acc)
	} else {
		return json.NewEncoder(w).Encode(lib)
	}
}

func (s *LibServer) GetLastBooks(w http.ResponseWriter, r *http.Request) error {
	account, err := readJWT(r, s.store)
	if err != nil {
		return err
	}
	books, err := s.store.GetLastBooks(int(account.ID))
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, books)
}

func (s *LibServer) GetAllLibrariesHandler(w http.ResponseWriter, r *http.Request) error {
	libraries, err := s.store.GetLibraries()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, libraries)
}

func (s *LibServer) GetSomeBooksHandler(w http.ResponseWriter, r *http.Request) error {
	books, err := s.store.GetSomeBooks()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, books)
}

type LibFunc func(w http.ResponseWriter, r *http.Request) error

type LibError struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	if status != http.StatusOK {
		fmt.Println("WriteJSON: ", v)
	}
	return json.NewEncoder(w).Encode(v)
}

func withJWTAuth(handlerFunc http.HandlerFunc, s database.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling JWT auth middleware")
		tokenString, err := r.Cookie("x-jwt-token")
		if err != nil {
			utils.PermissionDenied(w)
			return
		}
		token, err := validateJWT(tokenString.Value)
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
		if account.Email != claims["email"] {
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
	tokenString, err := r.Cookie("x-jwt-token")
	if err != nil {
		return nil, ErrorUnauthorized
	}
	token, err := validateJWT(tokenString.Value)
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
	account, err := s.GetAccountByEmail(claims["email"].(string))
	if err != nil {
		return nil, ErrorUnauthorized
	}

	return account, nil
}

func readLibJWT(r *http.Request, s database.Storage) (*types.LibraryAccount, error) {
	tokenString, err := r.Cookie("x-jwt-token")
	if err != nil {
		return nil, ErrorUnauthorized
	}
	token, err := validateJWT(tokenString.Value)
	if err != nil || !token.Valid {
		return nil, ErrorUnauthorized
	}
	claims := token.Claims.(jwt.MapClaims)
	if claims["email"] == nil {
		return nil, ErrorUnauthorized
	}
	lib, err := s.GetLibraryByEmail(claims["email"].(string))
	if err != nil {
		return nil, ErrorUnauthorized
	}
	return lib, err
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
		"expiresAt": time.Now().Add(time.Minute * 5).Unix(),
		"email":     email,
	}

	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func deleteJWT(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:    "x-jwt-token",
		Value:   "",
		Expires: time.Now(),
	}
	http.SetCookie(w, &cookie)
}
