package main

import (
	"./thisis"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"goji.io"
	. "goji.io/pat"
	"io/ioutil"
	"log"
	"net/http"
)

var INDEX string

func Index(w http.ResponseWriter, r *http.Request) {
	IndexLoad()
	fmt.Fprint(w, INDEX)
}

func CreateShortLink(w http.ResponseWriter, r *http.Request) {
	link :=
		thisis.Link{
			r.FormValue("originalurl"),
			Param(r, "lnk"),
			r.FormValue("changekey"),
		}

	res := thisis.IsShortLinkExist(link)

	switch res {
	case thisis.Exist:
		res := thisis.CheckLinkPassword(link)
		if res == thisis.WrongPassword {
			w.WriteHeader(http.StatusUnauthorized)
		} else if res == thisis.Success {
			thisis.UpdateShortLink(link)
			w.WriteHeader(http.StatusOK)
		} else {
			log.Panic(res)
		}
		break
	case thisis.NotExist:
		thisis.AddShortLink(link)
		w.WriteHeader(http.StatusCreated)
		break
	case thisis.DatabaseFailure:
		log.Print("DB ERROR")
		w.WriteHeader(http.StatusInternalServerError)
		break
	}
}

func DeleteShortLink(w http.ResponseWriter, r *http.Request) {
	link := thisis.Link{
		"",
		Param(r, "lnk"),
		r.FormValue("changekey"),
	}
	res := thisis.IsShortLinkExist(link)
	if res == thisis.Exist {
		res = thisis.CheckLinkPassword(link)
		if res == thisis.Success {
			thisis.DeleteShortLink(link)
			w.WriteHeader(http.StatusOK)
		} else if res == thisis.WrongPassword {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatal("Bcrypt Fault on URL /" + link.ShortUrl)
		}
	} else if res == thisis.NotExist {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal("Database Fault on URL /" + link.ShortUrl)
	}
}

func GetShortLink(w http.ResponseWriter, r *http.Request) {
	link := thisis.Link{
		"",
		Param(r, "lnk"),
		"",
	}
	res := thisis.GetShortLink(&link)
	if res == thisis.Success {
		http.Redirect(w, r, link.OriginalURL, http.StatusFound)
	} else if res == thisis.NotExist {
		http.NotFound(w, r)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal("Database Fault on URL /" + link.ShortUrl)
	}
}

func CheckShortLink(w http.ResponseWriter, r *http.Request) {
	shurl := Param(r, "lnk")
	res := thisis.IsShortLinkExist(thisis.Link{"", shurl, ""})
	switch res {
	case thisis.Exist:
		w.WriteHeader(http.StatusOK)
		break
	case thisis.NotExist:
		w.WriteHeader(http.StatusNoContent)
		break
	case thisis.DatabaseFailure:
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal("Database Fault on URL /" + shurl)
		break
	}
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
	DSN := "thisis:thisispasswd@/thisis"
	//flag.String()
	thisis.ConnectToDB(DSN)
	defer thisis.DB.Close()
	thisis.InitDB()
	//reg := regexp.MustCompile(`^/(?P<lnk>[A-Za-z0-9]+)$`)
	mux := goji.NewMux()
	mux.HandleFunc(Get("/"), Index)
	mux.HandleFunc(Get("/:file.:ext"), http.NotFound)
	mux.HandleFunc(Put("/:lnk"), CreateShortLink)
	mux.HandleFunc(Delete("/:lnk"), DeleteShortLink)
	mux.HandleFunc(Get("/:lnk"), GetShortLink)
	mux.HandleFunc(Post("/:lnk"), GetShortLink)
	mux.HandleFunc(Head("/:lnk"), CheckShortLink)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
