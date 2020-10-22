package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"net/url"
	"os"

	qrcode "github.com/skip2/go-qrcode"
)

type urlMaker interface {
	getURL(*url.URL) (string, error)
}

type urlFromID string

func (wp urlFromID) getURL(u *url.URL) (string, error) {
	id, ok := u.Query()["id"]
	if !ok || len(id[0]) == 0 {
		return "", errors.New("no ID found")
	}
	return string(wp) + "/send?id=" + id[0], nil
}

type urlFromURL struct{}

func (wp urlFromURL) getURL(u *url.URL) (string, error) {
	urlAddr, ok := u.Query()["url"]
	if !ok || len(urlAddr[0]) == 0 {
		return "", errors.New("no ID found")
	}
	return urlAddr[0], nil
}

func qrCodeHandler(u urlMaker) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		qrURL, err := u.getURL(r.URL)
		if err != nil {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		q, err := qrcode.New(qrURL, qrcode.Medium)
		if err != nil {
			http.Error(w, "Server Error", http.StatusInternalServerError)
			return
		}
		q.DisableBorder = true
		png, err := q.PNG(256)
		if err != nil {
			http.Error(w, "Server Error", http.StatusInternalServerError)
			return
		}
		log.Printf("Generated QR Code containing: %s", qrURL)
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Pragma", "public")
		w.Header().Set("Cache-Control", "max-age=86400")
		w.Write(png)
	}
}

func main() {
	port := flag.String("p", "3333", "port to serve")
	host := flag.String("h", "", "Server host")
	flag.Parse()

	serverPort := os.Getenv("SERVER_PORT")
	if len(serverPort) == 0 {
		serverPort = *port
	}

	serverHost := os.Getenv("SERVER_HOST")
	if len(serverPort) == 0 {
		serverHost = *host
	}

	var mf urlMaker
	if len(serverHost) > 0 {
		mf = urlFromID(serverHost)
	} else {
		mf = urlFromURL{}
	}

	http.HandleFunc("/qrcode", qrCodeHandler(mf))

	log.Printf("Serving on port %s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
