package main

import (
	"fmt"
	"github.com/Party4Bread/thisis"
	_ "github.com/go-sql-driver/mysql"
	"goji.io"
	. "goji.io/pat"
	"io/ioutil"
	"log"
	"net/http"
)

var INDEX string

func IndexLoad() {
	tINDEX, err := ioutil.ReadFile("index.html")
	if err != nil {
		log.Fatal("Index There is No IndexFile")
	}
	INDEX = string(tINDEX)
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, INDEX)
}
func main() {
	IndexLoad()
	DSN := "thisis:thisispasswd@/thisis"
	//flag.String()
	//reg := regexp.MustCompile(`^/(?P<lnk>[A-Za-z0-9]+)$`)
	mux := goji.NewMux()
	mux.HandleFunc(Get("/"), Index)
	thisis.SetupServer(mux, DSN)
	defer thisis.DB.Close()
	log.Fatal(http.ListenAndServe(":8080", mux))
}
