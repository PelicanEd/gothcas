package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/pelicaned/gothcas"
	"gopkg.in/cas.v2"
)

func initialize() {
	cookieKey := make([]byte, 12)
	_, err := rand.Read(cookieKey)
	if err != nil {
		log.Fatal(err)
	}

	casUcb, err := gothcas.New("https://casserver.herokuapp.com/", "http://localhost:8080/auth/callback", &gothcas.AttributeMap{
		Email:     "email",
		FirstName: "first-name",
		LastName:  "last-name",
		UserID:    "uid",
	})
	if err != nil {
		log.Fatal(err)
	}
	goth.UseProviders(casUcb)
}

func main() {
	initialize()

	mux := http.DefaultServeMux
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), "provider", "cas"))
		gothic.BeginAuthHandler(w, r)
	})
	mux.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			log.Print(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, user)
	})

	http.ListenAndServe("localhost:8080", nil)
}

func handle(w http.ResponseWriter, r *http.Request) {
	if !cas.IsAuthenticated(r) {
		cas.RedirectToLogin(w, r)
		return
	}

	if r.URL.Path == "/logout" {
		cas.RedirectToLogout(w, r)
		return
	}

	attr := cas.Attributes(r)

	fmt.Fprintln(w, attr)
}
