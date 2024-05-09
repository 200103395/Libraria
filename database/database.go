package database

import (
	"Libraria/types"
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"math/rand"
	"os"
	"strings"
	"time"
)

var (
	dbuser string
	dbpass string
	dbname string
	dbport string
	dbhost string
)

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
	OneTimeClear()
	CreateBook(book *types.Book) error
	GetBooks() (*[]types.Book, error)
	GetSomeBooks() (*[]types.Book, error)
	GetBookByID(id int) (*types.Book, error)
	UpdateBook(book types.Book) error
	DeleteBookByID(id int) error
	SearchBookName(name string) (*[]types.Book, *[]types.LibraryWeb, error)
	CreatePasswordReset(request *types.PasswordResetRequest) error
	GetPasswordReset(token string) (*types.PasswordResetRequest, error)
	DeletePasswordReset(request *types.PasswordResetRequest) error
	GetLibrariesByBookID(id int) (*[]types.LibraryAccount, error)
	GetBooksByLibraryID(id int) (*[]types.Book, error)
	AddBookVisit(user_id, book_id int)
	GetLastBooks(id int) (*[]types.Book, error)
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
	dbname = os.Getenv("DBNAME")
	dbhost = os.Getenv("DBHOST")
	dbport = os.Getenv("DBPORT")
	fmt.Println(dbuser, dbpass)
	//time.Sleep(5 * time.Second)
	connect := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", dbuser, dbpass, dbname, dbhost, dbport)
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

func (s *PostgresStorage) AddBookVisit(user_id, book_id int) {
	query := `select count(*) from last_books where user_id = $1 and book_id = $2`
	var count int
	if err := s.DB.QueryRow(query, user_id, book_id).Scan(&count); err != nil {
		return
	}
	timeNow := time.Now().UTC()
	if count == 0 {
		query = `insert into last_books(user_id, book_id, time) values ($1, $2, $3)`
		s.DB.Exec(query, user_id, book_id, timeNow)
		return
	} else {
		query = `update last_books set time = $3 where user_id = $1 and book_id = $2`
		s.DB.Exec(query, user_id, book_id, timeNow)
		return
	}
}

func (s *PostgresStorage) CreateBook(book *types.Book) error {
	query := `insert into book (name, author, year, genre, description, language, page_number) values ($1, $2, $3, $4, $5, $6);`
	_, err := s.DB.Exec(query, book.Name, book.Author, book.Year, book.Genre, book.Description, book.Language)
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
		err = rows.Scan(book.Pointers())
		if err != nil {
			continue
		}
		books = append(books, book)
	}
	return &books, nil
}

func (s *PostgresStorage) GetSomeBooks() (*[]types.Book, error) {
	var books []types.Book
	query := `select id,name,author,year,genre,description from book join 
(select * from (select count(library_id) as cnt, book_id from book_lib 
				group by book_id)) on book.id = book_id order by cnt desc limit 15;`
	rows, err := s.DB.Query(query)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	for rows.Next() {
		var book types.Book
		err = rows.Scan(book.Pointers())
		if err != nil {
			fmt.Println(err)
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
	err = row.Scan(book.Pointers())
	return &book, err
}

func (s *PostgresStorage) SearchBookName(name string) (*[]types.Book, *[]types.LibraryWeb, error) {
	var books []types.Book
	query := `select * from (
select id,name,author,year,genre,description from book join 
(select * from (select count(library_id) as cnt, book_id from book_lib 
				group by book_id)) on book.id = book_id order by cnt desc)
where lower(name) like $1 or lower(author) like $1 or lower(genre) like $1;`
	name = "%" + strings.ToLower(name) + "%"
	rows, err := s.DB.Query(query, name)
	if err != nil {
		return nil, nil, err
	}
	for rows.Next() {
		var book types.Book
		if err := rows.Scan(book.Pointers()); err != nil {
			fmt.Println(err)
			continue
		}
		books = append(books, book)
	}
	if len(books) == 0 {
		return nil, nil, fmt.Errorf("no books found")
	}
	query = `select id, email, name, address, contactnumber, latitude, longitude from library join
	(select distinct(library_id) from book_lib join (
    	select id from book where lower(name) like $1 or lower(author) like $1 or lower(genre) like $1
	) as sub on book_id = sub.id) as bigQuery on id = library_id;`
	var libs []types.LibraryWeb
	if rows, err = s.DB.Query(query, name); err != nil {
		return &books, nil, nil
	}
	for rows.Next() {
		var lib types.LibraryWeb
		if err = rows.Scan(&lib.ID, &lib.Email, &lib.Name, &lib.Address, &lib.ContactNumber, &lib.Latitude, &lib.Longitude); err != nil {
			fmt.Println(err)
			continue
		}
		libs = append(libs, lib)
	}
	return &books, &libs, nil
}

func (s *PostgresStorage) GetLastBooks(id int) (*[]types.Book, error) {
	var books []types.Book
	query := `select * from book join (
    	select distinct(book_id), time from last_books where user_id = $1
    	order by time desc limit 5 
	) as subquery on id = book_id`
	row, err := s.DB.Query(query, id)
	if err != nil {
		return nil, err
	}
	for row.Next() {
		var book types.Book
		var temp int
		var tim time.Time
		if err = row.Scan(&book.ID, &book.Name, &book.Author, &book.Year, &book.Genre, &book.Description, &temp, &tim); err != nil {
			return nil, err
		}
		books = append(books, book)
	}
	if len(books) == 0 {
		return nil, fmt.Errorf("No books found")
	}
	return &books, nil
}

func (s *PostgresStorage) GetBooksByLibraryID(id int) (*[]types.Book, error) {
	var books []types.Book
	query := `select id, name, author, year, genre, description 
	from book join (
		select distinct(book_id) from book_lib
		where library_id = $1
	) as sub on id = book_id;`
	rows, err := s.DB.Query(query, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var book types.Book
		if err = rows.Scan(book.Pointers()); err != nil {
			continue
		}
		books = append(books, book)
	}
	if len(books) == 0 {
		return nil, fmt.Errorf("no books found")
	}
	return &books, nil
}

func (s *PostgresStorage) GetLibrariesByBookID(id int) (*[]types.LibraryAccount, error) {
	var libs []types.LibraryAccount
	query := `select id, name, email, password, address, contactnumber, latitude, longitude
	from library join (
		select distinct(library_id) from book_lib 
		where book_id = $1
	) as subquery on id = library_id;`
	rows, err := s.DB.Query(query, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var lib types.LibraryAccount
		if err = rows.Scan(lib.Pointers()); err != nil {
			continue
		}
		libs = append(libs, lib)
	}
	if len(libs) == 0 {
		return nil, fmt.Errorf("no libraries found")
	}
	return &libs, nil
}

func (s *PostgresStorage) DeleteBookByID(id int) error {
	query := `delete from book where id = $1`
	_, err := s.DB.Exec(query, id)
	return err
}

func (s *PostgresStorage) UpdateBook(book types.Book) error {
	query := `update book set name = $1, author = $2, year = $3, genre = $4, description = $5, language = $6 where id = $7`
	_, err := s.DB.Exec(query, book.Name, book.Author, book.Year, book.Genre, book.Description, book.Language, book.ID)
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
