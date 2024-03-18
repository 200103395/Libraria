package types

import (
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserRequest struct {
	ID            uint      `json:"id"`
	FirstName     string    `json:"firstName"`
	LastName      string    `json:"lastName"`
	Password      string    `json:"password"`
	Email         string    `json:"email"`
	Address       string    `json:"address"`
	ContactNumber string    `json:"contactNumber"`
	Tag           string    `json:"tag"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type LibRequest struct {
	ID            uint      `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"libraryName"`
	Password      string    `json:"password"`
	Address       string    `json:"address"`
	ContactNumber string    `json:"contactNumber"`
	Tag           string    `json:"tag"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type Account struct {
	ID            uint   `json:"id"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Password      string `json:"password"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	ContactNumber string `json:"contactNumber"`
}

type LibraryAccount struct {
	ID            uint   `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"libraryName"`
	Password      string `json:"password"`
	Address       string `json:"address"`
	ContactNumber string `json:"contactNumber"`
}

type LibraryWeb struct {
	ID            uint   `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"libraryName"`
	Address       string `json:"address"`
	ContactNumber string `json:"contactNumber"`
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
	ID          uint     `json:"id"`
	Name        string   `json:"name"`
	Author      string   `json:"author"`
	Year        int      `json:"year"`
	Genre       []string `json:"genre"`
	Description string   `json:"description"`
	Language    string   `json:"language"`
	PageNumber  uint     `json:"pageNumber"`
}

type LibraryBook struct {
	BookID    uint `json:"bookID"`
	LibraryID uint `json:"libraryID"`
	Amount    uint `json:"amount"`
}

type Borrow struct {
	BookID     uint      `json:"bookID"`
	AccountID  uint      `json:"accountID"`
	LibraryID  uint      `json:"libraryID"`
	BorrowedAt time.Time `json:"borrowedAt"`
}

func (lib *LibraryAccount) ConvertToWeb() (webLibs *LibraryWeb) {
	return &LibraryWeb{
		ID:            lib.ID,
		Email:         lib.Email,
		Name:          lib.Name,
		Address:       lib.Address,
		ContactNumber: lib.ContactNumber,
	}
}

func (account *Account) Pointers() (*uint, *string, *string, *string, *string, *string, *string) {
	return &account.ID, &account.FirstName, &account.LastName, &account.Password, &account.Email, &account.Address, &account.ContactNumber
}

func (account *UserRequest) Pointers() (*uint, *string, *string, *string, *string, *string, *string, *string, *time.Time) {
	return &account.ID, &account.FirstName, &account.LastName, &account.Password, &account.Email, &account.Address, &account.ContactNumber, &account.Tag, &account.ExpiresAt
}

func (library *LibraryAccount) Pointers() (*uint, *string, *string, *string, *string, *string) {
	return &library.ID, &library.Name, &library.Email, &library.Password, &library.Address, &library.ContactNumber
}

func (library *LibRequest) Pointers() (*uint, *string, *string, *string, *string, *string, *string, *time.Time) {
	return &library.ID, &library.Name, &library.Email, &library.Password, &library.Address, &library.ContactNumber, &library.Tag, &library.ExpiresAt
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

func (account *Account) ValidateAccount() error {
	// Check all account credentials for validity
	// Password length and content
	// Name length and alphas
	// Email for validity
	return nil
}
func (library *LibraryAccount) ValidateAccount() error {
	// Check all account credentials for validity
	// Password length and content
	// Name length and alphas
	// Email for validity
	return nil
}
