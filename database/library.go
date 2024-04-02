package database

import (
	"Libraria/types"
	"database/sql"
	"fmt"
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
	query := "Insert into library (name, email, password, address, contactnumber, latitude, logitude) values ($1, $2, $3, $4, $5, $6, $7, $8);"
	_, err := s.DB.Exec(query, library.Name, library.Email, library.Password, library.Address, library.ContactNumber, library.Latitude, library.Longitude)
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
	query := `update library set name = $1, 
                   address = $2, contactNumber = $3, latitude = $4, longitude = $5 where id = $6;`
	_, err := s.DB.Exec(query, account.Name, account.Address, account.ContactNumber, account.Latitude, account.Longitude, account.ID)
	fmt.Println("HERE IS AN ERROR", err)
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
