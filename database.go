package main

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"os"
	"time"
)

var dbuser string
var dbpass string

type Storage interface {
	CheckEmail(Email string) (bool, error)
	CreateAccount(account *Account) error
	GetAccountByEmail(Email string) (*Account, error)
	GetAccountByID(id int) (*Account, error)
	GetAccounts() (*[]Account, error)
	UpdateAccount(account *Account) error
	CreateLibraryAccount(library *LibraryAccount) error
	GetLibraryByEmail(Email string) (*LibraryAccount, error)
	GetLibraryByID(id int) (*LibraryAccount, error)
	UpdateLibrary(account *LibraryAccount) error
	GetRequestByMAIL(mail string) (*CreateRequest, error)
	CreateRequest(request *CreateRequest) error
	DeleteRequest(request *CreateRequest) error
	ClearRequests()
	CreateBook(book *Book) error
	GetBooks() (*[]Book, error)
	GetBookByID(id int) (*Book, error)
	DeleteBookByID(id int) error
	CreatePasswordReset(request *PasswordResetRequest) error
	GetPasswordReset(token string) (*PasswordResetRequest, error)
	DeletePasswordReset(request *PasswordResetRequest) error
}

type PostgresStorage struct {
	DB *sql.DB
}

func NewPostgresStorage() (*PostgresStorage, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}
	dbuser = os.Getenv("DBUSERNAME")
	dbpass = os.Getenv("DBPASSWORD")
	fmt.Println(dbuser, dbpass)
	connect := fmt.Sprintf("user=%s password=%s dbname=postgres sslmode=disable", dbuser, dbpass)
	DB, err := sql.Open("postgres", connect)
	if err != nil {
		return nil, err
	}
	err = DB.Ping()
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{
		DB: DB,
	}, nil
}

func (s *PostgresStorage) Init() error {
	return s.CreateTables()
}

func (s *PostgresStorage) CheckEmail(Email string) (bool, error) {
	tables := []string{"account", "library", "requests"}
	for t := 0; t < len(tables); t++ {
		var count int
		query := "select count(*) from " + tables[t] + " where email = $1"
		err := s.DB.QueryRow(query, Email).Scan(&count)
		if err != nil {
			return false, err
		}
		if count > 0 {
			return false, nil
		}
	}
	return true, nil
}

// ###########################################################################################
func (s *PostgresStorage) CreateAccount(account *Account) error {
	query := "Insert into account (firstname, lastname, email, password, address, contactnumber) values ($1, $2, $3, $4, $5, $6);"
	_, err := s.DB.Exec(query, account.FirstName, account.LastName, account.Email, account.Password, account.Address, account.ContactNumber)
	return err
}

func (s *PostgresStorage) GetAccounts() (*[]Account, error) {
	rows, err := s.DB.Query("select * from account;")
	if err != nil {
		return nil, err
	}
	var accounts []Account
	for rows.Next() {
		var account Account
		if err = rows.Scan(account.Pointers()); err != nil {
			continue
		}
		accounts = append(accounts, account)
	}
	if len(accounts) == 0 {
		return nil, sql.ErrNoRows
	}
	return &accounts, err
}

func (s *PostgresStorage) GetAccountByID(id int) (*Account, error) {
	query := "select * from account where id = $1;"
	res, err := s.DB.Query(query, id)
	if err != nil {
		return nil, err
	}
	var account Account
	res.Next()
	err = res.Scan(account.Pointers())
	fmt.Printf("%+v\n", account)
	return &account, err
}

func (s *PostgresStorage) GetAccountByEmail(Email string) (*Account, error) {
	query := "select * from account where email = $1 limit 1"
	res := s.DB.QueryRow(query, Email)
	var account Account
	err := res.Scan(account.Pointers())
	return &account, err
}

func (s *PostgresStorage) UpdateAccount(account *Account) error {
	query := `update account set firstName = $1, lastName = $2, password = $3, 
                   address = $4, contactNumber = $5 where id = $6;`
	_, err := s.DB.Exec(query, account.FirstName, account.LastName, account.Password, account.Address, account.ContactNumber, account.ID)
	return err
}

// ###########################################################################################
func (s *PostgresStorage) CreateLibraryAccount(library *LibraryAccount) error {
	query := "Insert into library (name, email, password, address, contactnumber) values ($1, $2, $3, $4, $5, $6);"
	_, err := s.DB.Exec(query, library.Name, library.Email, library.Password, library.Address, library.ContactNumber)
	return err
}

func (s *PostgresStorage) GetLibraryByID(id int) (*LibraryAccount, error) {
	query := "select * from library where id = $1;"
	res, err := s.DB.Query(query, id)
	if err != nil {
		return nil, err
	}
	var library LibraryAccount
	res.Next()
	err = res.Scan(library.Pointers())
	fmt.Printf("%+v\n", library)
	return &library, err
}

func (s *PostgresStorage) GetLibraryByEmail(Email string) (*LibraryAccount, error) {
	query := "select * from library where email = $1"
	res, err := s.DB.Query(query, Email)
	if err != nil {
		return nil, err
	}
	var library LibraryAccount
	res.Next()
	err = res.Scan(library.Pointers())
	return &library, err
}

func (s *PostgresStorage) UpdateLibrary(account *LibraryAccount) error {
	query := `update account set name = $1, password = $2, 
                   address = $3, contactNumber = $4 where id = $6;`
	_, err := s.DB.Exec(query, account.Name, account.Password, account.Address, account.ContactNumber, account.ID)
	return err
}

// ###########################################################################################
func (s *PostgresStorage) CreateRequest(request *CreateRequest) error {
	query := "Insert into requests (firstname, lastname, email, password, address, contactnumber, tag, expires_at) values ($1, $2, $3, $4, $5, $6, $7, $8);"
	_, err := s.DB.Exec(query, request.FirstName, request.LastName, request.Email, request.Password, request.Address, request.ContactNumber, request.Tag, request.ExpiresAt)
	return err
}

func (s *PostgresStorage) GetRequestByMAIL(mail string) (*CreateRequest, error) {
	query := "select * from requests where email = $1;"
	res, err := s.DB.Query(query, mail)
	if err != nil {
		return nil, err
	}
	var req CreateRequest
	res.Next()
	err = res.Scan(req.Pointers())
	fmt.Printf("%+v\n", req)
	return &req, err
}

func (s *PostgresStorage) DeleteRequest(request *CreateRequest) error {
	query := "delete from requests where id = $1;"
	_, err := s.DB.Exec(query, request.ID)
	fmt.Println(err)
	return err
}

func (s *PostgresStorage) CreateBook(book *Book) error {
	query := `insert into book (name, author, year, genre, description, language, page_number) values ($1, $2, $3, $4, $5, $6, $7);`
	_, err := s.DB.Exec(query, book.Name, book.Author, book.Year, book.Genre, book.Description, book.Language, book.PageNumber)
	return err
}

func (s *PostgresStorage) GetBooks() (*[]Book, error) {
	var books []Book
	query := `select * from book;`
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var book Book
		err = rows.Scan(&book)
		if err != nil {
			continue
		}
		books = append(books, book)
	}
	return &books, nil
}

func (s *PostgresStorage) GetBookByID(id int) (*Book, error) {
	var book Book
	query := `select * from book where id = $1`
	row, err := s.DB.Query(query, id)
	if err != nil {
		return &book, err
	}
	row.Next()
	err = row.Scan(&book)
	return &book, err
}

func (s *PostgresStorage) DeleteBookByID(id int) error {
	query := `delete from book where id = $1`
	_, err := s.DB.Exec(query, id)
	return err
}

func (s *PostgresStorage) CreatePasswordReset(request *PasswordResetRequest) error {
	query := `insert into password_reset (email, token, expires_at) values ($1, $2, $3)`
	_, err := s.DB.Exec(query, request.Email, request.Token, request.ExpiresAt)
	return err
}

func (s *PostgresStorage) GetPasswordReset(token string) (*PasswordResetRequest, error) {
	query := `select * from password_reset where token = $1 limit 1;`
	row := s.DB.QueryRow(query, token)
	var req PasswordResetRequest
	err := row.Scan(&req.ID, &req.Email, &req.Token, &req.ExpiresAt)
	return &req, err
}

func (s *PostgresStorage) DeletePasswordReset(request *PasswordResetRequest) error {
	query := `delete from password_reset where id = $1;`
	_, err := s.DB.Exec(query, request.ID)
	return err
}

func (s *PostgresStorage) CreateTables() error {
	query := `CREATE TABLE IF NOT EXISTS account (
    id SERIAL PRIMARY KEY,
    firstname VARCHAR(50),
    lastname VARCHAR(50),
    password VARCHAR(200),
    email VARCHAR(50),
    address VARCHAR(200),
    contactnumber VARCHAR(20));`
	_, err := s.DB.Exec(query)
	if err != nil {
		return err
	}

	query = `create table if not exists library(
    id SERIAL PRIMARY KEY,
    name varchar(200),
    email varchar(50),
    password varchar(200),
    address varchar(200),
    contactnumber varchar(20)
)`
	_, err = s.DB.Exec(query)
	if err != nil {
		return err
	}

	query = `create table if not exists book(
    id SERIAL PRIMARY KEY,
    name varchar(255),
    author varchar(255),
    year INT NOT NULL,
    genre TEXT[] NOT NULL,
    description TEXT,
    language VARCHAR(255),
    page_number INT
)`
	_, err = s.DB.Exec(query)
	if err != nil {
		return err
	}

	query = `create table if not exists password_reset(
    id SERIAL PRIMARY KEY,
    email varchar(255),
    token varchar(255),
    expires_at TIMESTAMP NOT NULL
)`
	_, err = s.DB.Exec(query)
	if err != nil {
		return err
	}

	query = `CREATE TABLE IF NOT EXISTS requests (
    id SERIAL PRIMARY KEY,
    firstname VARCHAR(50),
    lastname VARCHAR(50),
    password VARCHAR(200),
    email VARCHAR(50),
    address VARCHAR(200),
    contactnumber VARCHAR(20),
    tag VARCHAR(65),
    expires_at TIMESTAMP NOT NULL);`
	_, err = s.DB.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) DropTable(name string) error {
	query := "drop table %s"
	_, err := s.DB.Exec(fmt.Sprintf(query, name))
	return err
}

func (s *PostgresStorage) ClearRequests() {
	fmt.Println("Clearing requests has been started")
	for {
		rows, err := s.DB.Query(`select id from requests where expires_at < NOW() at time zone 'UTC';`)
		if err != nil {
			fmt.Println("Error while clearing requests:", err)
			continue
		}
		for rows.Next() {
			var id int
			err = rows.Scan(&id)
			if err != nil {
				fmt.Println("Error while iterating requests", err)
				continue
			}
			_, err = s.DB.Exec(`delete from requests where id = $1`, id)
			if err != nil {
				fmt.Println("Error while deleting requests", err)
				continue
			}
			fmt.Println("Request with id ", id, " was deleted")
		}
		rows.Close()

		rows, err = s.DB.Query(`select id from password_reset where expires_at < NOW() at time zone 'UTC';`)
		if err != nil {
			fmt.Println("Error while clearing requests:", err)
			continue
		}
		for rows.Next() {
			var id int
			err = rows.Scan(&id)
			if err != nil {
				fmt.Println("Error while iterating requests", err)
				continue
			}
			_, err = s.DB.Exec(`delete from password_reset where id = $1`, id)
			if err != nil {
				fmt.Println("Error while deleting requests", err)
				continue
			}
			fmt.Println("Request with id ", id, " was deleted")
		}
		rows.Close()

		time.Sleep(1 * time.Minute)
	}
}
