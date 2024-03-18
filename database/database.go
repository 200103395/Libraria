package database

import (
	"Libraria/types"
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"math/rand"
	"os"
)

var dbuser string
var dbpass string

type Storage interface {
	CheckEmail(Email string) (bool, error)
	CheckForRequest(email string) error
	MakeToken(Table string) string
	CreateAccount(account *types.Account) error
	GetAccountByEmail(Email string) (*types.Account, error)
	GetAccountByID(id int) (*types.Account, error)
	UpdateAccount(account *types.Account) error
	GetLibraries() (*[]types.LibraryWeb, error)
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
	tables := []string{"account", "library", "user_requests", "lib_requests"}
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

func (s *PostgresStorage) MakeToken(Table string) string {
	var alphaNumRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	emailVerRandRunes := make([]rune, 64)
	for true {
		for i := 0; i < 64; i++ {
			emailVerRandRunes[i] = alphaNumRunes[rand.Intn(len(alphaNumRunes)-1)]
		}
		var count int
		query := "select count(*) from " + Table + " where tag = $1"
		if err := s.DB.QueryRow(query, string(emailVerRandRunes)).Scan(&count); err != nil {
			continue
		}
		if count == 0 {
			break
		}
	}
	return string(emailVerRandRunes)
}

// ###########################################################################################
func (s *PostgresStorage) CheckForRequest(email string) error {
	query := "select count(*) from password_reset where email = $1"
	var count int
	err := s.DB.QueryRow(query, email).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("request exists")
	}
	return nil
}

func (s *PostgresStorage) CreateUserRequest(request *types.UserRequest) error {
	query := "Insert into user_requests (firstname, lastname, email, password, address, contactnumber, tag, expires_at) values ($1, $2, $3, $4, $5, $6, $7, $8);"
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
	query := "Insert into lib_requests (name, email, password, address, contactnumber, tag, expires_at) values ($1, $2, $3, $4, $5, $6, $7);"
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
	query := `insert into password_reset (email, tag, expires_at) values ($1, $2, $3)`
	_, err := s.DB.Exec(query, request.Email, request.Token, request.ExpiresAt)
	return err
}

func (s *PostgresStorage) GetPasswordReset(token string) (*types.PasswordResetRequest, error) {
	query := `select * from password_reset where tag = $1 limit 1;`
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
