package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"goji.io"
	"goji.io/pat"
	"golang.org/x/crypto/bcrypt"
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
	shurl, passwd := pat.Param(r, "lnk"), r.FormValue("changekey")
	originURL := r.FormValue("originalurl")

	var keyhash string
	err := db.QueryRow("SELECT ChangeKey FROM URLstorage WHERE ShortedURL=?",
		shurl).Scan(&keyhash)

	if err != nil {
		if err == sql.ErrNoRows {
			phash, err := bcrypt.GenerateFromPassword([]byte(passwd), 14)
			if err != nil {
				log.Fatal("bcrypt failure")
			}
			db.Query("INSERT INTO urlstorage (ShortedURL, OriginalURL, ChangeKey) VALUES (?,?,?)",
				shurl, originURL, phash)
			w.WriteHeader(http.StatusCreated)
		} else {
			log.Fatal("DB ERROR")
		}
	} else {
		if bcrypt.CompareHashAndPassword([]byte(keyhash), []byte(passwd)) == nil {
			db.Query("UPDATE urlstorage SET OriginalURL=? WHERE ShortedURL=?",
				originURL, shurl)
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}
}

func DeleteShortLink(w http.ResponseWriter, r *http.Request) {

}

func GetShortLink(w http.ResponseWriter, r *http.Request) {
	var originalUrl string
	shurl := pat.Param(r, "lnk")
	err := db.QueryRow("SELECT OriginalURL FROM URLstorage WHERE ShortedURL=?",
		shurl).Scan(&originalUrl)
	if err != nil {
		if err == sql.ErrNoRows {
			//TODO: Add Link not Founded
		} else {
			//TODO: Add Database Error
			log.Fatal("Database Fault on URL /" + shurl)
		}
	} else {
		http.Redirect(w, r, originalUrl, http.StatusFound)
	}
}

func CheckShortLink(w http.ResponseWriter, r *http.Request) {
	shurl := pat.Param(r, "lnk")
	err := db.QueryRow("SELECT OriginalURL FROM URLstorage WHERE ShortedURL=?",
		shurl).Scan()
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatal("Database Fault on URL /" + shurl)
		}
	} else {
		w.WriteHeader(http.StatusOK)
	}
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
