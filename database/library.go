package database

import (
	"Libraria/types"
	"database/sql"
)

func (s *PostgresStorage) GetLibraries() (*[]types.LibraryWeb, error) {
	rows, err := s.DB.Query("select * from library;")
	if err != nil {
		return nil, err
	}
	var accounts []types.LibraryWeb
	for rows.Next() {
		var library types.LibraryAccount
		if err = rows.Scan(library.Pointers()); err != nil {
			continue
		}
		accounts = append(accounts, *library.ConvertToWeb())
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
