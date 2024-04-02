package controllers

import (
	"Libraria/types"
	"Libraria/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func (s *LibServer) SearchHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		html, err := os.ReadFile("static/search.html")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, string(html))
		return err
	}
	if r.Method != "POST" {
		return utils.MethodNotAllowed(w)
	}
	var inJson struct {
		Input string `json:"inputValue"`
	}
	if err := json.NewDecoder(r.Body).Decode(&inJson); err != nil {
		fmt.Println(err)
		return err
	}
	books, libs, err := s.store.SearchBookName(inJson.Input)
	if err != nil {
		return err
	}
	ret := struct {
		Books     []types.Book       `json:"books"`
		Libraries []types.LibraryWeb `json:"libraries"`
	}{
		Books:     *books,
		Libraries: *libs,
	}
	jsonBooks, err := json.Marshal(ret)
	if err != nil {
		return err
	}
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(jsonBooks)
	return err
}
