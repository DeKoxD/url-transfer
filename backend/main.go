package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/mediocregopher/radix"
)

func getID(r *http.Request) (string, error) {
	ipPort := r.RemoteAddr
	ip := rPort.ReplaceAllString(ipPort, "")
	if len(ip) < 3 {
		return "", errors.New("Invalid IP")
	}
	ua := r.UserAgent()
	if len(ip) < 3 {
		return "", errors.New("Invalid User Agent")
	}
	sum := sha512.Sum512_224([]byte(ip + ua))
	str :=
		strings.ReplaceAll(
			strings.ReplaceAll(
				base64.StdEncoding.EncodeToString(sum[:])[:16],
				"+", "_"),
			"/", "-")
	return str, nil
}

type redisURLStorage struct {
	pool   *radix.Pool
	expire string
}

func (r redisURLStorage) set(id string, urlStr string) error {
	if len(id) == 0 {
		return errors.New("Invalid ID")
	}
	var exists int
	if err := r.pool.Do(radix.Cmd(&exists, "EXISTS", id)); err != nil || exists == 0 {
		log.Printf("Error: redisURLStorage.set: Cmd SETEX: %s", err)
		return errors.New("ID expired or does not exist")
	}
	log.Println(id, exists) //------------------------------------------------------------------------------------------
	var setVal string
	if err := r.pool.Do(radix.Cmd(&setVal, "SETEX", id, r.expire, urlStr)); err != nil {
		log.Printf("Error: redisURLStorage.set: Cmd SETEX: %s", err)
		return errors.New("Couldn't set URL")
	}
	return nil
}

func (r redisURLStorage) get(id string) (string, error) {
	if len(id) == 0 {
		return "", errors.New("Invalid ID")
	}
	var getVal string
	mn := radix.MaybeNil{Rcv: &getVal}
	if err := r.pool.Do(radix.Cmd(&mn, "GET", id)); err != nil || mn.Nil {
		log.Printf("Error: redisURLStorage.get: Cmd GET: %s", err)
		return "", errors.New("Couldn't get URL")
	}
	return getVal, nil
}

func (r redisURLStorage) create(id string) error {
	if len(id) == 0 {
		return errors.New("Invalid ID")
	}
	var setVal string
	if err := r.pool.Do(radix.Cmd(&setVal, "SETNX", id, "")); err != nil {
		log.Printf("Error: redisURLStorage.func create: Cmd SETNX: %s", err)
		return errors.New("Couldn't set URL")
	}
	if err := r.pool.Do(radix.Cmd(&setVal, "EXPIRE", id, r.expire)); err != nil {
		log.Printf("Error: redisURLStorage.create: Cmd EXPIRE: %s", err)
		return errors.New("Couldn't set URL")
	}
	return nil
}

// URLResponse is a JSON template to respond to '/url' requests
type URLResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

var rURL = regexp.MustCompile(`/url/([a-zA-Z0-9_-]+)={0,2}`)

func getIDFromURL(path string) (string, error) {
	aID := rURL.FindStringSubmatch(path)
	if len(aID) != 2 {
		return "", errors.New("Invalid URL")
	}
	if len(aID[1]) == 0 {
		return "", errors.New("Invalid ID")
	}
	return aID[1], nil
}

func urlHandler(u *redisURLStorage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := getIDFromURL(r.URL.Path)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case "GET":
			su, err := u.get(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			data, err := json.Marshal(&URLResponse{ID: id, URL: su})
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		case "PUT":
			var pl URLResponse
			err := json.NewDecoder(r.Body).Decode(&pl)
			if err != nil {
				http.Error(w, "Invalid Request", http.StatusBadRequest)
				return
			}
			if len(pl.URL) > 2047 {
				pl.URL = pl.URL[:2047]
			}
			err = u.set(id, pl.URL)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
	}
}

func allowCORS(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	return w
}

var rPort = regexp.MustCompile(":[0-9]{2,6}$")

// RegisterResponse is a JSON template to respond to '/register' requests
type RegisterResponse struct {
	ID string `json:"id"`
}

func registerHandler(u *redisURLStorage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := getID(r)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusUnauthorized)
			return
		}
		data, err := json.Marshal(&RegisterResponse{ID: id})
		if err != nil {
			http.Redirect(w, r, "/", http.StatusUnauthorized)
			return
		}
		err = u.create(id)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		log.Println(id)
		w.Write(data)
	}
}

func main() {
	port := flag.String("p", "3333", "port to serve")
	redis := flag.String("r", "db:6379", "redis database address with port")
	expire := flag.Int("e", 3600, "redis key timeout")
	flag.Parse()

	serverPort := os.Getenv("SERVER_PORT")
	if len(serverPort) == 0 {
		serverPort = *port
	}
	redisAddr := os.Getenv("REDIS_ADDR")
	if len(redisAddr) == 0 {
		redisAddr = *redis
	}
	redisExpire := os.Getenv("REDIS_EXPIRE")
	if rex, err := strconv.Atoi(redisExpire); rex <= 0 || err != nil {
		if *expire > 0 {
			redisExpire = strconv.Itoa(*expire)
		} else {
			redisExpire = "3600"
		}
	}

	pool, err := radix.NewPool("tcp", redisAddr, 10)
	if err != nil {
		log.Panic(err)
	}

	ru := &redisURLStorage{pool: pool, expire: redisExpire}
	http.HandleFunc("/url/", urlHandler(ru))
	http.HandleFunc("/register", registerHandler(ru))

	log.Printf("Serving on port %s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
