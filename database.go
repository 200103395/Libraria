package main

import (
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"os"
)

var dbuser string
var dbpass string

type Storage interface {
	CreateAccount(account *Account) error
	CreateLibraryAccount(library *LibraryAccount) error
	GetAccountByEmail(Email string) (*Account, error)
	GetAccountByID(id int) (*Account, error)
	GetAccounts() (*[]Account, error)
	GetLibraryByEmail(Email string) (*LibraryAccount, error)
	GetLibraryByID(id int) (*LibraryAccount, error)
	CheckEmail(Email string) (bool, error)
	GetRequestByMAIL(mail string) (*CreateRequest, error)
	CreateRequest(request *CreateRequest) error
	DeleteRequest(request *CreateRequest) error
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

func (s *PostgresStorage) CreateAccount(account *Account) error {
	query := "Insert into account (firstname, lastname, email, password, address, contactnumber) values ($1, $2, $3, $4, $5, $6);"
	_, err := s.DB.Exec(query, account.FirstName, account.LastName, account.Email, account.Password, account.Address, account.ContactNumber)
	return err
}

func (s *PostgresStorage) CreateRequest(request *CreateRequest) error {
	query := "Insert into requests (firstname, lastname, email, password, address, contactnumber, tag) values ($1, $2, $3, $4, $5, $6, $7);"
	_, err := s.DB.Exec(query, request.FirstName, request.LastName, request.Email, request.Password, request.Address, request.ContactNumber, request.Tag)
	return err
}

func (s *PostgresStorage) CreateLibraryAccount(library *LibraryAccount) error {
	query := "Insert into library (name, email, password, address, contactnumber) values ($1, $2, $3, $4, $5, $6);"
	_, err := s.DB.Exec(query, library.Name, library.Email, library.Password, library.Address, library.ContactNumber)
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
	query := "select * from account where email = $1"
	res, err := s.DB.Query(query, Email)
	if err != nil {
		return nil, err
	}
	var account Account
	res.Next()
	err = res.Scan(account.Pointers())
	return &account, err
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

func (s *PostgresStorage) DropAccount() error {
	query := `drop table account;`
	_, err := s.DB.Exec(query)
	return err
}

func (s *PostgresStorage) DropBook() error {
	query := `drop table book;`
	_, err := s.DB.Exec(query)
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
    name varchar(200),
    author varchar(100),
    year SERIAL,
    description varchar(200),
    cover varchar(200)
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
    tag VARCHAR(65));`
	_, err = s.DB.Exec(query)
	if err != nil {
		return err
	}
	return nil
}
