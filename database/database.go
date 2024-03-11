package database

import (
	"Libraria/types"
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
	CreateAccount(account *types.Account) error
	GetAccountByEmail(Email string) (*types.Account, error)
	GetAccountByID(id int) (*types.Account, error)
	UpdateAccount(account *types.Account) error
	GetLibraries() (*[]types.LibraryAccount, error)
	CreateLibraryAccount(library *types.LibraryAccount) error
	GetLibraryByEmail(Email string) (*types.LibraryAccount, error)
	GetLibraryByID(id int) (*types.LibraryAccount, error)
	UpdateLibrary(account *types.LibraryAccount) error
	GetUserRequestByTAG(tag string) (*types.UserRequest, error)
	CreateUserRequest(request *types.UserRequest) error
	DeleteUserRequest(request *types.UserRequest) error

	GetLibRequestByTAG(tag string) (*types.LibRequest, error)
	CreateLibRequest(request *types.LibRequest) error
	DeleteLibRequest(request *types.LibRequest) error
	ClearRequests()
	CreateBook(book *types.Book) error
	GetBooks() (*[]types.Book, error)
	GetBookByID(id int) (*types.Book, error)
	DeleteBookByID(id int) error
	CreatePasswordReset(request *types.PasswordResetRequest) error
	GetPasswordReset(token string) (*types.PasswordResetRequest, error)
	DeletePasswordReset(request *types.PasswordResetRequest) error
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
func (s *PostgresStorage) CreateAccount(account *types.Account) error {
	query := "Insert into account (firstname, lastname, email, password, address, contactnumber) values ($1, $2, $3, $4, $5, $6);"
	_, err := s.DB.Exec(query, account.FirstName, account.LastName, account.Email, account.Password, account.Address, account.ContactNumber)
	return err
}

func (s *PostgresStorage) GetAccountByID(id int) (*types.Account, error) {
	query := "select * from account where id = $1;"
	res, err := s.DB.Query(query, id)
	if err != nil {
		return nil, err
	}
	var account types.Account
	res.Next()
	err = res.Scan(account.Pointers())
	fmt.Printf("%+v\n", account)
	return &account, err
}

func (s *PostgresStorage) GetAccountByEmail(Email string) (*types.Account, error) {
	query := "select * from account where email = $1 limit 1"
	res := s.DB.QueryRow(query, Email)
	var account types.Account
	err := res.Scan(account.Pointers())
	return &account, err
}

func (s *PostgresStorage) UpdateAccount(account *types.Account) error {
	query := `update account set firstName = $1, lastName = $2, password = $3, 
                   address = $4, contactNumber = $5 where id = $6;`
	_, err := s.DB.Exec(query, account.FirstName, account.LastName, account.Password, account.Address, account.ContactNumber, account.ID)
	return err
}

// ###########################################################################################

func (s *PostgresStorage) GetLibraries() (*[]types.LibraryAccount, error) {
	rows, err := s.DB.Query("select * from library;")
	if err != nil {
		return nil, err
	}
	var accounts []types.LibraryAccount
	for rows.Next() {
		var library types.LibraryAccount
		if err = rows.Scan(library.Pointers()); err != nil {
			continue
		}
		accounts = append(accounts, library)
	}
	if len(accounts) == 0 {
		return nil, sql.ErrNoRows
	}
	return &accounts, err
}

func (s *PostgresStorage) CreateLibraryAccount(library *types.LibraryAccount) error {
	query := "Insert into library (name, email, password, address, contactnumber) values ($1, $2, $3, $4, $5, $6);"
	_, err := s.DB.Exec(query, library.Name, library.Email, library.Password, library.Address, library.ContactNumber)
	return err
}

func (s *PostgresStorage) GetLibraryByID(id int) (*types.LibraryAccount, error) {
	query := "select * from library where id = $1;"
	res, err := s.DB.Query(query, id)
	if err != nil {
		return nil, err
	}
	var library types.LibraryAccount
	res.Next()
	err = res.Scan(library.Pointers())
	fmt.Printf("%+v\n", library)
	return &library, err
}

func (s *PostgresStorage) GetLibraryByEmail(Email string) (*types.LibraryAccount, error) {
	query := "select * from library where email = $1"
	res, err := s.DB.Query(query, Email)
	if err != nil {
		return nil, err
	}
	var library types.LibraryAccount
	res.Next()
	err = res.Scan(library.Pointers())
	return &library, err
}

func (s *PostgresStorage) UpdateLibrary(account *types.LibraryAccount) error {
	query := `update account set name = $1, password = $2, 
                   address = $3, contactNumber = $4 where id = $6;`
	_, err := s.DB.Exec(query, account.Name, account.Password, account.Address, account.ContactNumber, account.ID)
	return err
}

// ###########################################################################################
func (s *PostgresStorage) CreateUserRequest(request *types.UserRequest) error {
	query := "Insert into requests (firstname, lastname, email, password, address, contactnumber, tag, expires_at) values ($1, $2, $3, $4, $5, $6, $7, $8);"
	_, err := s.DB.Exec(query, request.FirstName, request.LastName, request.Email, request.Password, request.Address, request.ContactNumber, request.Tag, request.ExpiresAt)
	return err
}

func (s *PostgresStorage) GetUserRequestByTAG(tag string) (*types.UserRequest, error) {
	query := "select * from user_requests where tag = $1;"
	res, err := s.DB.Query(query, tag)
	if err != nil {
		return nil, err
	}
	var req types.UserRequest
	res.Next()
	err = res.Scan(req.Pointers())
	fmt.Printf("%+v\n", req)
	return &req, err
}

func (s *PostgresStorage) DeleteUserRequest(request *types.UserRequest) error {
	query := "delete from user_requests where id = $1;"
	_, err := s.DB.Exec(query, request.ID)
	fmt.Println(err)
	return err
}

func (s *PostgresStorage) CreateLibRequest(request *types.LibRequest) error {
	query := "Insert into requests (name, email, password, address, contactnumber, tag, expires_at) values ($1, $2, $3, $4, $5, $6, $7);"
	_, err := s.DB.Exec(query, request.Name, request.Email, request.Password, request.Address, request.ContactNumber, request.Tag, request.ExpiresAt)
	return err
}

func (s *PostgresStorage) GetLibRequestByTAG(tag string) (*types.LibRequest, error) {
	query := "select * from lib_requests where tag = $1;"
	res, err := s.DB.Query(query, tag)
	if err != nil {
		return nil, err
	}
	var req types.LibRequest
	res.Next()
	err = res.Scan(req.Pointers())
	fmt.Printf("%+v\n", req)
	return &req, err
}

func (s *PostgresStorage) DeleteLibRequest(request *types.LibRequest) error {
	query := "delete from lib_requests where id = $1;"
	_, err := s.DB.Exec(query, request.ID)
	fmt.Println(err)
	return err
}

func (s *PostgresStorage) CreateBook(book *types.Book) error {
	query := `insert into book (name, author, year, genre, description, language, page_number) values ($1, $2, $3, $4, $5, $6, $7);`
	_, err := s.DB.Exec(query, book.Name, book.Author, book.Year, book.Genre, book.Description, book.Language, book.PageNumber)
	return err
}

func (s *PostgresStorage) GetBooks() (*[]types.Book, error) {
	var books []types.Book
	query := `select * from book;`
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var book types.Book
		err = rows.Scan(&book)
		if err != nil {
			continue
		}
		books = append(books, book)
	}
	return &books, nil
}

func (s *PostgresStorage) GetBookByID(id int) (*types.Book, error) {
	var book types.Book
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

func (s *PostgresStorage) CreatePasswordReset(request *types.PasswordResetRequest) error {
	query := `insert into password_reset (email, token, expires_at) values ($1, $2, $3)`
	_, err := s.DB.Exec(query, request.Email, request.Token, request.ExpiresAt)
	return err
}

func (s *PostgresStorage) GetPasswordReset(token string) (*types.PasswordResetRequest, error) {
	query := `select * from password_reset where token = $1 limit 1;`
	row := s.DB.QueryRow(query, token)
	var req types.PasswordResetRequest
	err := row.Scan(&req.ID, &req.Email, &req.Token, &req.ExpiresAt)
	return &req, err
}

func (s *PostgresStorage) DeletePasswordReset(request *types.PasswordResetRequest) error {
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

	query = `CREATE TABLE IF NOT EXISTS user_requests (
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

	query = `CREATE TABLE IF NOT EXISTS lib_requests (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50),
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
		rows, err := s.DB.Query(`select id from user_requests where expires_at < NOW() at time zone 'UTC';`)
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
			_, err = s.DB.Exec(`delete from user_requests where id = $1`, id)
			if err != nil {
				fmt.Println("Error while deleting requests", err)
				continue
			}
			fmt.Println("Request with id ", id, " was deleted")
		}
		rows.Close()

		rows, err = s.DB.Query(`select id from lib_requests where expires_at < NOW() at time zone 'UTC';`)
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
			_, err = s.DB.Exec(`delete from lib_requests where id = $1`, id)
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
