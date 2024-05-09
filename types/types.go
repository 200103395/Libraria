package types

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"time"
	"unicode"
)

type UserRequest struct {
	ID        uint      `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Password  string    `json:"password"`
	Email     string    `json:"email"`
	Tag       string    `json:"tag"`
	ExpiresAt time.Time `json:"expires_at"`
}

type LibRequest struct {
	ID            uint      `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	Password      string    `json:"password"`
	Address       string    `json:"address"`
	ContactNumber string    `json:"contactNumber"`
	Latitude      float32   `json:"latitude"`
	Longitude     float32   `json:"longitude"`
	Tag           string    `json:"tag"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type Account struct {
	ID        uint   `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Password  string `json:"password"`
	Email     string `json:"email"`
}

type LibraryAccount struct {
	ID            uint    `json:"id"`
	Email         string  `json:"email"`
	Name          string  `json:"name"`
	Password      string  `json:"password"`
	Address       string  `json:"address"`
	ContactNumber string  `json:"contactNumber"`
	Latitude      float32 `json:"latitude"`
	Longitude     float32 `json:"longitude"`
}

type LibraryWeb struct {
	ID            uint    `json:"id"`
	Email         string  `json:"email"`
	Name          string  `json:"name"`
	Address       string  `json:"address"`
	ContactNumber string  `json:"contactNumber"`
	Latitude      float32 `json:"latitude"`
	Longitude     float32 `json:"longitude"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PasswordResetRequest struct {
	ID        uint      `json:"ID"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type NewPassword struct {
	NewPassword        string `json:"newPassword"`
	NewPasswordConfirm string `json:"newPasswordConfirm"`
}

type Book struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Author      string `json:"author"`
	Year        uint   `json:"year"`
	Genre       string `json:"genre"`
	Description string `json:"description"`
	Language    string `json:"language"`
}

type LibraryBook struct {
	BookID    uint `json:"bookID"`
	LibraryID uint `json:"libraryID"`
	Amount    uint `json:"amount"`
}

func (lib *LibraryAccount) ConvertToWeb() (webLibs *LibraryWeb) {
	return &LibraryWeb{
		ID:            lib.ID,
		Email:         lib.Email,
		Name:          lib.Name,
		Address:       lib.Address,
		ContactNumber: lib.ContactNumber,
		Latitude:      lib.Latitude,
		Longitude:     lib.Longitude,
	}
}

func (book *Book) Pointers() (*uint, *string, *string, *uint, *string, *string) {
	return &book.ID, &book.Name, &book.Author, &book.Year, &book.Genre, &book.Description
}

func (account *Account) Pointers() (*uint, *string, *string, *string, *string) {
	return &account.ID, &account.FirstName, &account.LastName, &account.Password, &account.Email
}

func (account *UserRequest) Pointers() (*uint, *string, *string, *string, *string, *string, *time.Time) {
	return &account.ID, &account.FirstName, &account.LastName, &account.Password, &account.Email, &account.Tag, &account.ExpiresAt
}

func (library *LibraryAccount) Pointers() (*uint, *string, *string, *string, *string, *string, *float32, *float32) {
	return &library.ID, &library.Name, &library.Email, &library.Password, &library.Address, &library.ContactNumber, &library.Latitude, &library.Longitude
}

func (library *LibRequest) Pointers() (*uint, *string, *string, *string, *string, *string, *float32, *float32, *string, *time.Time) {
	return &library.ID, &library.Name, &library.Email, &library.Password, &library.Address, &library.ContactNumber, &library.Latitude, &library.Longitude, &library.Tag, &library.ExpiresAt
}

func (account *Account) ValidPassword(pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(pw)) == nil
}

func (library *LibraryAccount) ValidPassword(pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(library.Password), []byte(pw)) == nil
}

func (account *Account) PasswordHash() error {
	encpw, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	account.Password = string(encpw)
	return nil
}

func (library *LibraryAccount) PasswordHash() error {
	encpw, err := bcrypt.GenerateFromPassword([]byte(library.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	library.Password = string(encpw)
	return nil
}

func containsLetters(s string) bool {
	for _, char := range s {
		if !unicode.IsLetter(char) {
			return false
		}
	}
	return true
}

func containsLetterAndDigits(s string) bool {
	for _, char := range s {
		if !unicode.IsDigit(char) && !unicode.IsLetter(char) {
			return false
		}
	}
	return true
}

func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(email)
}

func isValidNumber(number string) bool {
	pattern := `^\+\d{11}$`
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(number)
}

func (account *Account) ValidateAccount() error {
	if len(account.Email) == 0 || len(account.Email) > 50 || !isValidEmail(account.Email) {
		return fmt.Errorf("invalid email")
	}
	if len(account.Password) < 8 || len(account.Password) > 20 || !containsLetterAndDigits(account.Password) {
		return fmt.Errorf("invalid password")
	}
	if len(account.FirstName) == 0 || len(account.FirstName) > 15 || !containsLetters(account.FirstName) {
		return fmt.Errorf("invalid first name")
	}
	if len(account.LastName) == 0 || len(account.LastName) > 15 || !containsLetters(account.LastName) {
		return fmt.Errorf("invalid last name")
	}
	return nil
}
func (library *LibraryAccount) ValidateAccount() error {
	if len(library.Email) == 0 || len(library.Email) > 50 || !isValidEmail(library.Email) {
		return fmt.Errorf("invalid email")
	}
	if len(library.Password) < 8 || len(library.Password) > 20 || !containsLetterAndDigits(library.Password) {
		return fmt.Errorf("invalid password")
	}
	if len(library.Name) == 0 || len(library.Name) > 50 || !containsLetters(library.Name) {
		return fmt.Errorf("invalid first name")
	}
	if !isValidNumber(library.ContactNumber) {
		return fmt.Errorf("invalid phone number")
	}
	if len(library.Address) == 0 || len(library.Address) > 50 || !containsLetterAndDigits(library.Address) {
		return fmt.Errorf("invalid address")
	}
	if library.Latitude < -90.0 || library.Latitude > 90.0 {
		return fmt.Errorf("invalid latitude")
	}
	if library.Longitude < -180.0 || library.Longitude > 180.0 {
		return fmt.Errorf("invalid longitude")
	}
	return nil
}
