package thisis

import (
	"goji.io"
	"goji.io/pat"
	"log"
	"net/http"
)

func ShortLinkCreate(w http.ResponseWriter, r *http.Request) {
	link :=
		Link{
			r.FormValue("originalurl"),
			pat.Param(r, "lnk"),
			r.FormValue("changekey"),
		}

	res := IsShortLinkExist(link)

	switch res {
	case Exist:
		res := CheckLinkPassword(link)
		if res == WrongPassword {
			w.WriteHeader(http.StatusUnauthorized)
		} else if res == Success {
			UpdateShortLink(link)
			w.WriteHeader(http.StatusOK)
		} else {
			log.Panic(res)
		}
		break
	case NotExist:
		AddShortLink(link)
		w.WriteHeader(http.StatusCreated)
		break
	case DatabaseFailure:
		log.Print("DB ERROR")
		w.WriteHeader(http.StatusInternalServerError)
		break
	}
}

func ShortLinkDelete(w http.ResponseWriter, r *http.Request) {
	link := Link{
		"",
		pat.Param(r, "lnk"),
		r.FormValue("changekey"),
	}
	res := IsShortLinkExist(link)
	if res == Exist {
		res = CheckLinkPassword(link)
		if res == Success {
			DeleteShortLink(link)
			w.WriteHeader(http.StatusOK)
		} else if res == WrongPassword {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatal("Bcrypt Fault on URL /" + link.ShortUrl)
		}
	} else if res == NotExist {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal("Database Fault on URL /" + link.ShortUrl)
	}
}

func ShortLinkGet(w http.ResponseWriter, r *http.Request) {
	link := Link{
		"",
		pat.Param(r, "lnk"),
		"",
	}
	res := GetShortLink(&link)
	if res == Success {
		http.Redirect(w, r, link.OriginalURL, http.StatusFound)
	} else if res == NotExist {
		http.NotFound(w, r)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal("Database Fault on URL /" + link.ShortUrl)
	}
}

func CheckShortLink(w http.ResponseWriter, r *http.Request) {
	shurl := pat.Param(r, "lnk")
	res := IsShortLinkExist(Link{"", shurl, ""})
	switch res {
	case Exist:
		w.WriteHeader(http.StatusOK)
		break
	case NotExist:
		w.WriteHeader(http.StatusNoContent)
		break
	case DatabaseFailure:
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal("Database Fault on URL /" + shurl)
		break
	}
}

func SetupServer(mux goji.Mux, DSN string) {

	ConnectToDB(DSN)
	defer DB.Close()
	InitDB()

	mux.HandleFunc(pat.Get("/:file.:ext"), http.NotFound)
	mux.HandleFunc(pat.Put("/:lnk"), ShortLinkCreate)
	mux.HandleFunc(pat.Delete("/:lnk"), ShortLinkDelete)
	mux.HandleFunc(pat.Get("/:lnk"), ShortLinkGet)
	mux.HandleFunc(pat.Post("/:lnk"), ShortLinkGet)
	mux.HandleFunc(pat.Head("/:lnk"), CheckShortLink)
}
