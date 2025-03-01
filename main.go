package main

import (
	"fmt"
	"log"
	"net/http"
)

//endpoint --> /shorten (POST) --> cikti olarak kisa olan url'i donecegiz.
//endpoint --> /xYxa (kisa url) --> redirection yapacagiz ve siteye dogrudan gidecek.

//service-->
//database -->




func main(){
	http.HandleFunc("/shorten", shortenUrlHandler)
	http.HandleFunc("/:shortUrl", redirectUrlHandler)


	log.Fatal(http.ListenAndServe(":8080", nil))
}