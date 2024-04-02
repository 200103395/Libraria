package database

import (
	"Libraria/types"
	"fmt"
)

func (s *PostgresStorage) CreateAccount(account *types.Account) error {
	query := "Insert into account (firstname, lastname, email, password) values ($1, $2, $3, $4);"
	_, err := s.DB.Exec(query, account.FirstName, account.LastName, account.Email, account.Password)
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
	query := `update account set firstName = $1, lastName = $2 where id = $3;`
	_, err := s.DB.Exec(query, account.FirstName, account.LastName, account.ID)
	return err
}

func (s *PostgresStorage) CreateUserRequest(request *types.UserRequest) error {
	query := "Insert into user_requests (firstname, lastname, email, password, tag, expires_at) values ($1, $2, $3, $4, $5, $6);"
	_, err := s.DB.Exec(query, request.FirstName, request.LastName, request.Email, request.Password, request.Tag, request.ExpiresAt)
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
