package main

import (
	"golang.org/x/crypto/bcrypt"
	"time"
)

type CreateRequest struct {
	ID            uint   `json:"id"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Password      string `json:"password"`
	Email         string `json:"email"`
	Address       string `json:"address"`
	ContactNumber string `json:"contactNumber"`
	Tag           string `json:"tag"`
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

func (account *Account) Pointers() (*uint, *string, *string, *string, *string, *string, *string) {
	return &account.ID, &account.FirstName, &account.LastName, &account.Password, &account.Email, &account.Address, &account.ContactNumber
}

func (account *CreateRequest) Pointers() (*uint, *string, *string, *string, *string, *string, *string, *string) {
	return &account.ID, &account.FirstName, &account.LastName, &account.Password, &account.Email, &account.Address, &account.ContactNumber, &account.Tag
}

func (library *LibraryAccount) Pointers() (*uint, *string, *string, *string, *string, *string) {
	return &library.ID, &library.Name, &library.Email, &library.Password, &library.Address, &library.ContactNumber
}

type LibraryAccount struct {
	ID            uint   `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"libraryName"`
	Password      string `json:"password"`
	Address       string `json:"address"`
	ContactNumber string `json:"contactNumber"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (account *Account) ValidPassword(pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(pw)) == nil
}

func (library *LibraryAccount) ValidPassword(pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(library.Password), []byte(pw)) == nil
}

func NewAccount(account *Account) error {
	encpw, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	account.Password = string(encpw)

	return nil
}

func NewLibraryAccount(library *LibraryAccount) error {
	encpw, err := bcrypt.GenerateFromPassword([]byte(library.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	library.Password = string(encpw)

	return nil
}

func NewAccountVars(firstName, lastName, password, email, address, contact string) (*Account, error) {
	encpw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &Account{
		FirstName:     firstName,
		LastName:      lastName,
		Password:      string(encpw),
		Email:         email,
		Address:       address,
		ContactNumber: contact,
	}, nil
}

type Book struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Author      string `json:"author"`
	Year        int    `json:"year"`
	Description string `json:"description"`
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
