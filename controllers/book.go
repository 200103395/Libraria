package controllers

import (
	"Libraria/types"
	"Libraria/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func (s *LibServer) GetLibrariesByBookIDHandler(w http.ResponseWriter, r *http.Request) error {
	id, err := utils.GetID(r)
	if err != nil {
		return err
	}
	libs, err := s.store.GetLibrariesByBookID(id)
	if err != nil {
		return err
	}
	var web_libs []types.LibraryWeb
	for i := 0; i < len(*libs); i++ {
		web := (*libs)[i].ConvertToWeb()
		web_libs = append(web_libs, *web)
	}
	return WriteJSON(w, http.StatusOK, web_libs)
}

func (s *LibServer) GetBooksByLibraryIDHandler(w http.ResponseWriter, r *http.Request) error {
	id, err := utils.GetID(r)
	if err != nil {
		return err
	}
	books, err := s.store.GetBooksByLibraryID(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, books)
}

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
	if err != nil {
		return err
	}
	bookJson, err := json.Marshal(book)
	if err != nil {
		return err
	}
	html, err := os.ReadFile("static/bookGet.html")
	if err != nil {
		return err
	}
	if _, err = fmt.Fprintf(w, fmt.Sprintf("<script>var book = %s;</script>", bookJson)); err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, string(html))
	if acc, err := readJWT(r, s.store); err == nil {
		fmt.Println("Adding a book", acc)
		s.store.AddBookVisit(int(acc.ID), id)
	}
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
