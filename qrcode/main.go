package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	qrcode "github.com/skip2/go-qrcode"
)

func qrCodeHandler(w http.ResponseWriter, r *http.Request) {
	qrURL, ok := r.URL.Query()["url"]
	if !ok || len(qrURL[0]) == 0 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
	}

	q, err := qrcode.New(qrURL[0], qrcode.Medium)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
	}
	q.DisableBorder = true
	png, err := q.PNG(256)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
	}
	log.Printf("Generated QR Code containing: %s", qrURL)
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Pragma", "public")
	w.Header().Set("Cache-Control", "max-age=86400")
	w.Write(png)
}

func main() {
	port := flag.String("p", "3333", "port to serve")
	flag.Parse()

	serverPort := os.Getenv("SERVER_PORT")
	if len(serverPort) == 0 {
		serverPort = *port
	}

	http.HandleFunc("/qrcode", qrCodeHandler)

	log.Printf("Serving on port %s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
