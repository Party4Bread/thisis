package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"goji.io"
	. "goji.io/pat"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"log"
	"net/http"
)

var db *sql.DB

const DSN = "thisis:thisispasswd@/thisis"

var INDEX string

func Index(w http.ResponseWriter, r *http.Request) {
	IndexLoad()
	fmt.Fprint(w, INDEX)
}
func CreateShortLink(w http.ResponseWriter, r *http.Request) {
	shurl, passwd, originURL :=
		Param(r, "lnk"), r.FormValue("changekey"), r.FormValue("originalurl")

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
	shurl, passwd := Param(r, "lnk"), r.FormValue("changekey")

	var keyhash string
	err := db.QueryRow("SELECT ChangeKey FROM URLstorage WHERE ShortedURL=?",
		shurl).Scan(&keyhash)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatal("Database Fault on URL /" + shurl)
		}
	} else {
		if bcrypt.CompareHashAndPassword([]byte(keyhash), []byte(passwd)) == nil {
			db.Query("DELETE FROM URLstorage WHERE ShortedURL=?",
				shurl)
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}
}

func GetShortLink(w http.ResponseWriter, r *http.Request) {
	var originalUrl string
	shurl := Param(r, "lnk")
	err := db.QueryRow("SELECT OriginalURL FROM URLstorage WHERE ShortedURL=?",
		shurl).Scan(&originalUrl)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatal("Database Fault on URL /" + shurl)
		}
	} else {
		http.Redirect(w, r, originalUrl, http.StatusFound)
	}
}

func CheckShortLink(w http.ResponseWriter, r *http.Request) {
	shurl := Param(r, "lnk")
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

func IndexLoad() {
	tINDEX, err := ioutil.ReadFile("index.html")
	if err != nil {
		log.Fatal("Index There is No IndexFile")
	}
	INDEX = string(tINDEX)
}
func main() {
	IndexLoad()
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
	mux.HandleFunc(Get("/"), Index)
	mux.HandleFunc(Put("/:lnk"), CreateShortLink)
	mux.HandleFunc(Delete("/:lnk"), DeleteShortLink)
	mux.HandleFunc(Get("/:lnk"), GetShortLink)
	mux.HandleFunc(Post("/:lnk"), GetShortLink)
	mux.HandleFunc(Head("/:lnk"), CheckShortLink)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
