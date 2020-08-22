package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

var tpl = template.Must(template.ParseFiles("index.html"))

type Results struct {
	Word      string      `json:"word"`
	Phonetics []Phonetics `json:"phonetics"`
	Origin    string      `json:"origin"`
	Meaning   Meaning     `json:"meaning"`
}
type Phonetics struct {
	Text  string `json:"text"`
	Audio string `json:"audio"`
}
type Meaning struct {
	Noun []Noun `json:"noun"`
	Verb []Verb `json:"verb"`
}
type Noun struct {
	Definition string `json:"definition"`
	Example    string `json:"example"`
}
type Verb struct {
	Definition string   `json:"definition"`
	Example    string   `json:"example"`
	Synonyms   []string `json:"synonyms"`
}

type DictionaryError struct {
	// Status  string `json:"status"`
	// Code    string `json:"code"`
	Message string `json:"message"`
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.String())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	params := u.Query()
	searchKey := params.Get("q")
	if len(searchKey) == 0 {
		tpl.Execute(w, nil)
		return
	}

	url := fmt.Sprintf("https://api.dictionaryapi.dev/api/v1/entries/en/%s", searchKey)
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		newError := &DictionaryError{}
		err := json.NewDecoder(resp.Body).Decode(newError)
		if err != nil {
			http.Error(w, "Unexpected server error", http.StatusInternalServerError)
			return
		}

		http.Error(w, newError.Message, http.StatusInternalServerError)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)

	respBody := []byte(string(body))
	resultStruts := []Results{}

	json.Unmarshal(respBody, &resultStruts)
	err = tpl.Execute(w, resultStruts)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	mux.HandleFunc("/search", searchHandler)
	mux.HandleFunc("/", indexHandler)
	http.ListenAndServe(":"+port, mux)
}
