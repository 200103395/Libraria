package database

import (
	"fmt"
	"time"
)

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
    name varchar(400),
    author varchar(400),
    year INT NOT NULL,
    genre TEXT NOT NULL,
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
    tag varchar(255),
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
