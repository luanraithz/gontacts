package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	"log"

	"github.com/gorilla/pat"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

type Source struct {
	UpdateTime string `json:"updateTime"`
	Type       string `json:"type"`
}

type PhoneMetadata struct {
	Primary bool `json:"primary"`
}

type Name struct {
	DisplayName string `json:"displayName"`
}

type PhoneNumber struct {
	Value string `json:"value"`
}
type ContactMetadata struct {
	Sources []Source `json:"sources"`
}

type Photo struct {
	Default bool   `json:"default"`
	Url     string `json:"url"`
}
type Contact struct {
	Metadata ContactMetadata `json:"metadata"`
	Names    []Name          `json:"names"`
}

type Result struct {
	Connections []Contact `json:"connections"`
	TotalPeople int       `json:"totalPeople"`
}

func main() {
	url := os.Getenv("DOMAIN_URL")
	if url == "" {
		url = "http://localhost:3000"
	}
	goth.UseProviders(
		google.New(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), os.Getenv("DOMAIN_URL")+"/auth/google/callback", "email", "profile", "https://www.googleapis.com/auth/contacts.readonly"),
	)

	p := pat.New()
	p.Get("/auth/{provider}/callback", func(res http.ResponseWriter, req *http.Request) {

		user, err := gothic.CompleteUserAuth(res, req)
		if err != nil {
			http.Redirect(res, req, "/", http.StatusSeeOther)
			println(err)
			return
		}
		t, _ := template.ParseFiles("templates/success.html")
		url := `https://people.googleapis.com/v1/people/me/connections?sortOrder=LAST_MODIFIED_DESCENDING&personFields=metadata&personFields=names`
		reqContacts, _ := http.NewRequest("GET", url, nil)
		reqContacts.Header.Set("Authorization", "Bearer "+user.AccessToken)
		ans, err := http.DefaultClient.Do(reqContacts)
		if err != nil {
			http.Redirect(res, req, "/", http.StatusSeeOther)
			println(err)
			return
		}
		defer ans.Body.Close()
		body, err := ioutil.ReadAll(ans.Body)
		if err != nil {
			http.Redirect(res, req, "/", http.StatusSeeOther)
			println(err)
			return
		}
		result := Result{}
		json.Unmarshal(body, &result)

		t.Execute(res, map[string]interface{}{
			"Email":       user.Email,
			"Name":        user.Name,
			"Connections": result.Connections,
		})
	})

	p.Get("/logout/{provider}", func(res http.ResponseWriter, req *http.Request) {
		gothic.Logout(res, req)
		res.Header().Set("Location", "/")
		res.WriteHeader(http.StatusTemporaryRedirect)
	})

	p.Get("/auth/{provider}", func(res http.ResponseWriter, req *http.Request) {
		gothic.BeginAuthHandler(res, req)
	})

	p.PathPrefix("/static").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))
	p.Get("/login", func(res http.ResponseWriter, req *http.Request) {
		t, _ := template.ParseFiles("templates/login.html")
		t.Execute(res, false)
	})
	p.Get("/", func(res http.ResponseWriter, req *http.Request) {
		t, _ := template.ParseFiles("templates/index.html")
		t.Execute(res, false)
	})

	log.Println("listening on localhost:3000")
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), p))
}
