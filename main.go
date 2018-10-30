package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"goji.io"
	"goji.io/pat"
	"log"
	"net/http"
)

var db *sql.DB

const DSN = "thisis:thisispasswd@/thisis"

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello, %s!\n", pat.Param(r, "name"))
}

func CreateShortLink(w http.ResponseWriter, r *http.Request) {

}

func DeleteShortLink(w http.ResponseWriter, r *http.Request) {

}

func GetShortLink(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "http://party4bread.github.io/", 301)
}

func CheckShortLink(w http.ResponseWriter, r *http.Request) {

}

func InitDB() {
	db.Query(`CREATE TABLE IF NOT EXISTS URLStorage(
  ID           INT            NOT NULL    AUTO_INCREMENT,
  ShortedURL   VARCHAR(45),
  OriginalURL  TEXT,
  ChangeKey    VARCHAR(60),
  PRIMARY KEY (ID))`)
}

func main() {
	tdb, err := sql.Open("mysql", DSN)
	if err != nil {
		log.Fatal("DB Connection Failed")
		panic(err.Error())
	}
	defer tdb.Close()
	err = tdb.Ping()
	if err != nil {
		log.Fatal("DB Ping Failed")
		panic(err.Error())
	}

	db = tdb
	InitDB()

	mux := goji.NewMux()
	mux.HandleFunc(pat.Get("/"), Index)
	mux.HandleFunc(pat.Put("/:lnk"), CreateShortLink)
	mux.HandleFunc(pat.Delete("/:lnk"), DeleteShortLink)
	mux.HandleFunc(pat.Get("/:lnk"), GetShortLink)
	mux.HandleFunc(pat.Post("/:lnk"), GetShortLink)
	mux.HandleFunc(pat.Head("/:lnk"), CheckShortLink)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
