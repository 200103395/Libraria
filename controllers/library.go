package controllers

import (
	"Libraria/types"
	"Libraria/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func (s *LibServer) LibraryHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	libraries, err := s.store.GetLibraries()
	if err != nil {
		return err
	}
	html, err := os.ReadFile("static/library.html")
	if err != nil {
		return err
	}
	libJson, err := json.Marshal(libraries)
	if err != nil {
		return err
	}
	if _, err = fmt.Fprintf(w, "<script>var libraries = %s;</script>", libJson); err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, string(html))
	return err
}

func (s *LibServer) LibrarySettingsHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		if _, err := readLibJWT(r, s.store); err != nil {
			return err
		}
		html, err := os.ReadFile("static/librarySettings.html")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, string(html))
		return err
	}
	if r.Method != "POST" {
		fmt.Println(r.URL, r.Method)
		return utils.MethodNotAllowed(w)
	}
	lib, err := readLibJWT(r, s.store)
	fmt.Println("READING LIB", lib, err)
	if err != nil {
		return err
	}
	var newLib types.LibraryAccount
	if err := json.NewDecoder(r.Body).Decode(&newLib); err != nil {
		fmt.Println("Decoder error", err, newLib)
		return err
	}
	if err := newLib.ValidateAccount(); err != nil {
		return err
	}
	newLib.ID = lib.ID
	fmt.Println("LIB", newLib)
	if err := s.store.UpdateLibrary(&newLib); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, newLib)
}

func (s *LibServer) LibraryCreateHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" && r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	if r.Method == "GET" {
		html, err := os.ReadFile("static/libraryRegister.html")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, string(html))
		return err
		// return HTML file
	}
	s.store.OneTimeClear()
	var library types.LibraryAccount
	if err := json.NewDecoder(r.Body).Decode(&library); err != nil {
		return err
	}
	if err := library.ValidateAccount(); err != nil {
		return err
	}
	isExists, err := s.store.CheckEmail(library.Email)
	if err != nil {
		return err
	}
	if isExists == false {
		return WriteJSON(w, http.StatusBadRequest, fmt.Sprintf("Email %s is already in use", library.Email))
	}

	expiresAt := time.Now().Add(expiration * time.Minute).UTC()
	req := types.LibRequest{
		Name:          library.Name,
		Password:      library.Password,
		Email:         library.Email,
		Address:       library.Address,
		ContactNumber: library.ContactNumber,
		Latitude:      library.Latitude,
		Longitude:     library.Longitude,
		Tag:           s.store.MakeToken("lib_requests"),
		ExpiresAt:     expiresAt,
	}
	if err = s.store.CreateLibRequest(&req); err != nil {
		return err
	}
	appeal := req.Name + " Library"
	if err = s.email.EmailConfirmationMessage(req.Email, appeal, domain+"/account/confirm/"+req.Tag); err != nil {
		return err
	}
	ans := []any{fmt.Sprintf("Message was sent to %s , to confirm account please follow the instructions in the message", req.Email), req}
	return WriteJSON(w, http.StatusOK, ans)
	// show success HTML page

}

func (s *LibServer) LibraryConfirmHandler(w http.ResponseWriter, r *http.Request) error {
	s.store.OneTimeClear()
	if r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	tag := utils.GetTAG(r)
	lib, err := s.store.GetLibRequestByTAG(tag)
	if err != nil {
		return err
	}
	library := &types.LibraryAccount{
		Name:          lib.Name,
		Email:         lib.Email,
		Password:      lib.Password,
		Address:       lib.Address,
		ContactNumber: lib.ContactNumber,
		Latitude:      lib.Latitude,
		Longitude:     lib.Longitude,
	}
	err = library.PasswordHash()
	if err != nil {
		return err
	}
	err = s.store.CreateLibraryAccount(library)
	if err != nil {
		return err
	}
	err = s.store.DeleteLibRequest(lib)
	if err != nil {
		// DO SOMETHING
	}
	html, err := os.ReadFile("static/libraryConfirm.html")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, string(html))
	return err
}

func (s *LibServer) GetLibraryHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	id, err := utils.GetID(r)
	if err != nil {
		return err
	}
	lib, err := s.store.GetLibraryByID(id)
	if err != nil {
		return err
	}
	html, err := os.ReadFile("static/libraryGet.html")
	if err != nil {
		return err
	}
	jsonLib := lib.ConvertToWeb()
	jsonMar, err := json.Marshal(jsonLib)
	if err != nil {
		return err
	}
	if _, err = fmt.Fprintf(w, "<script id=\"headScript\">var lib = %s;</script>", jsonMar); err != nil {
		return err
	}
	if _, err = fmt.Fprintf(w, string(html)); err != nil {
		return err
	}
	return nil
}

func (s *LibServer) LibraryLoginHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" && r.Method != "GET" {
		return utils.MethodNotAllowed(w)
	}
	if r.Method == "GET" {
		html, err := os.ReadFile("static/libraryLogin.html")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, string(html))
		return err
	}

	var req types.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	lib, err := s.store.GetLibraryByEmail(req.Email)
	if err != nil {
		return err
	}

	if !lib.ValidPassword(req.Password) {
		return fmt.Errorf("not authenticated")
	}

	token, err := createJWT(lib.Email)
	if err != nil {
		return err
	}

	cookie := http.Cookie{
		Name:     "x-jwt-token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(time.Minute * loginTimeMinutes),
	}
	http.SetCookie(w, &cookie)
	return WriteJSON(w, http.StatusOK, cookie)
}
