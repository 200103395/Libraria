package controllers

import (
	"Libraria/types"
	"Libraria/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func (s *LibServer) BookHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "DELETE" {
		return s.BookDeleteHandler(w, r)
	}
	if r.Method == "UPDATE" {
		return s.BookUpdateHandler(w, r)
	}
	if r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	id, err := utils.GetID(r)
	if err != nil {
		return err
	}
	book, err := s.store.GetBookByID(id)
	bookJson, err := json.Marshal(book)
	if err != nil {
		return err
	}
	html, err := os.ReadFile("static/bookGet.html")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, fmt.Sprintf("<script>var book = %s;</script>", bookJson))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, string(html))
	return err
}

func (s *LibServer) BookDeleteHandler(w http.ResponseWriter, r *http.Request) error {
	id, err := utils.GetID(r)
	if err != nil {
		return err
	}
	err = s.store.DeleteBookByID(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, "Delete successful")
}

func (s *LibServer) BookUpdateHandler(w http.ResponseWriter, r *http.Request) error {
	id, err := utils.GetID(r)
	if err != nil {
		return err
	}
	var book types.Book
	err = json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		return err
	}
	book.ID = uint(id)
	err = s.store.UpdateBook(book)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, "Update successful")
}

func (s *LibServer) BookCreateHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		html, err := os.ReadFile("static/bookCreate.html")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, string(html))
		return err
	}
	if r.Method != "POST" {
		return utils.MethodNotAllowed(w)
	}
	var book types.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		return err
	}
	if err := s.store.CreateBook(&book); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, "Book successfully created")
}
