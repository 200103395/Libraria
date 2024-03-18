package database

import (
	"Libraria/types"
	"fmt"
)

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
