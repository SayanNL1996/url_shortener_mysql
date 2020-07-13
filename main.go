package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/ventu-io/go-shortid"
)

type Url struct {
	ShortURL string `json:"shorturl"`
	LongURL  string `json:"longurl"`
}

type Responsestruct struct {
	Message  string
	Response Url
}

var db *sql.DB
var err error

func createurl(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	shorturl, err := shortid.Generate()
	stmt, err := db.Prepare("INSERT INTO url_table(shorturl,longurl) VALUES(?,?)")
	if err != nil {
		panic(err.Error())
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err.Error())
	}
	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal)
	longurl := keyVal["longurl"]
	_, err = stmt.Exec(shorturl, longurl)
	if err != nil {
		panic(err.Error())
	}
	var responsestruct = Responsestruct{
		Message: "Short URL generated",
		Response: Url{
			LongURL:  longurl,
			ShortURL: r.Host + "/" + shorturl,
		},
	}
	jsonResponse, err := json.Marshal(responsestruct)
	w.Write(jsonResponse)

}

// This function will redirect the short url to the actual url

func Redirecturl(w http.ResponseWriter, r *http.Request) {

	shorturl := mux.Vars(r)["shorturl"]
	if shorturl == " " {
		fmt.Println("Error")
	} else {

		fmt.Println("short", shorturl)
		result, err := db.Query("SELECT longurl from url_table where shorturl = ?", shorturl)
		if err != nil {
			panic(err.Error())
		}
		defer result.Close()
		var url Url
		for result.Next() {
			err := result.Scan(&url.LongURL)
			if err != nil {
				panic(err.Error())
			}
			fmt.Println("result", url.LongURL)
		}
		http.Redirect(w, r, url.LongURL, http.StatusSeeOther)
		json.NewEncoder(w).Encode(url)
	}

}

func main() {
	db, err = sql.Open("mysql", "root:Qwerty@123@tcp(127.0.0.1:3306)/urls")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	router := mux.NewRouter()
	router.HandleFunc("/api/url", createurl).Methods("POST")
	router.HandleFunc("/{shorturl}", Redirecturl).Methods("GET")

	http.ListenAndServe(":8020", router)

}
