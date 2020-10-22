package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/mediocregopher/radix"
)

var pSalt string

func pSaltRoutine(period int) {
	var t time.Time
	st := time.Duration(period) * time.Second
	div := int(math.Ceil(float64(period) / 60))
	for {
		t = time.Now()

		pSalt = t.String()[:13] + strconv.Itoa(t.Minute()/div)
		time.Sleep(st)
	}
}

func getHash(r *http.Request) (string, error) {
	ipPort := r.RemoteAddr
	ip := rPort.ReplaceAllString(ipPort, "")
	if len(ip) < 3 {
		return "", errors.New("Invalid address")
	}
	ua := r.UserAgent()
	if len(ip) < 3 {
		return "", errors.New("Invalid User Agent")
	}
	sum := sha512.Sum512_224([]byte(ip + ua))
	str := base64.StdEncoding.EncodeToString(sum[:])
	return str, nil
}

func getID(hash string, salt string) string {
	sum := sha512.Sum512_224([]byte(salt + pSalt + hash))
	var id string
	for _, n := range sum[:9] {
		id += strconv.Itoa(int(n) % 10)
	}
	return id
}

type redisURLStorage struct {
	poolURL   *radix.Pool
	poolID    *radix.Pool
	expireURL string
	expireID  string
	timeout   int
}

func (r redisURLStorage) set(id string, urlStr string) error {
	if len(id) == 0 {
		return errors.New("Invalid ID")
	}
	var getValHash string
	mn := radix.MaybeNil{Rcv: &getValHash}
	if err := r.poolID.Do(radix.Cmd(&mn, "GET", id)); err != nil || mn.Nil {
		log.Printf("Error: redisURLStorage.set: ID Cmd GET: %s", err)
		return errors.New("ID expired or does not exist")
	}
	var exists int
	if err := r.poolURL.Do(radix.Cmd(&exists, "EXISTS", getValHash)); err != nil || exists == 0 {
		log.Printf("Error: redisURLStorage.set: URL Cmd EXISTS: %s", err)
		return errors.New("ID expired or does not exist")
	}
	var setVal string
	if err := r.poolURL.Do(radix.Cmd(&setVal, "SETEX", getValHash, r.expireURL, urlStr)); err != nil {
		log.Printf("Error: redisURLStorage.set: URL Cmd SETEX: %s", err)
		return errors.New("Couldn't set URL")
	}
	return nil
}

func (r redisURLStorage) get(id string, hash string) (string, error) {
	if len(id) == 0 {
		return "", errors.New("Invalid ID")
	}
	var getValHash string
	mn := radix.MaybeNil{Rcv: &getValHash}
	if err := r.poolID.Do(radix.Cmd(&mn, "GET", id)); err != nil || mn.Nil {
		log.Printf("Error: redisURLStorage.get: ID Cmd GET: %s", err)
		return "", errors.New("ID expired or does not exist")
	}
	if getValHash != hash {
		log.Printf("Error: redisURLStorage.get: Unauthorized user: %s != %s", getValHash, hash)
		return "", errors.New("Unauthorized")
	}
	var getVal string
	mn.Rcv = &getVal
	if err := r.poolURL.Do(radix.Cmd(&mn, "GET", getValHash)); err != nil || mn.Nil {
		log.Printf("Error: redisURLStorage.get: URL Cmd GET: %s", err)
		return "", errors.New("Couldn't get URL")
	}
	log.Printf("hash:%v, %v", hash, getVal)
	return getVal, nil
}

func (r redisURLStorage) create(hash string) (string, error) {
	if len(hash) == 0 {
		return "", errors.New("Invalid ID")
	}
	var id string
	var getValHash string
	var setVal string
	mn := radix.MaybeNil{Rcv: &getValHash}
	for i := 0; ; i++ {
		if i > 5 {
			return "", errors.New("Invalid ID")
		}
		id = getID(hash, strconv.Itoa(i))
		if err := r.poolID.Do(radix.Cmd(&mn, "GET", id)); err != nil {
			log.Printf("Error: redisURLStorage.func create: ID Cmd GET: %s", err)
			return "", errors.New("Couldn't check ID")
		}
		if mn.Nil {
			if err := r.poolID.Do(radix.Cmd(&setVal, "SETEX", id, r.expireID, hash)); err != nil {
				log.Printf("Error: redisURLStorage.func create: ID Cmd SETEX: %s", err)
				return "", errors.New("Couldn't register ID")
			}
			break
		} else if hash == getValHash {
			break
		}
	}
	if err := r.poolURL.Do(radix.Cmd(&setVal, "SETNX", hash, "")); err != nil {
		log.Printf("Error: redisURLStorage.func create: URL Cmd SETNX: %s", err)
		return "", errors.New("Couldn't register ID")
	}
	if err := r.poolURL.Do(radix.Cmd(&setVal, "EXPIRE", hash, r.expireURL)); err != nil {
		log.Printf("Error: redisURLStorage.create: URL Cmd EXPIRE: %s", err)
		return "", errors.New("Couldn't register ID")
	}
	return id, nil
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
			hash, err := getHash(r)
			if err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
			su, err := u.get(id, hash)
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
	ID      string `json:"id"`
	Timeout int    `json:"timeout"`
}

func registerHandler(u *redisURLStorage) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		hash, err := getHash(r)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusUnauthorized)
			return
		}
		id, err := u.create(hash)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError) // TODO: Change message
			return
		}
		data, err := json.Marshal(&RegisterResponse{ID: id, Timeout: u.timeout})
		if err != nil {
			http.Redirect(w, r, "/", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		log.Println(id)
		w.Write(data)
	}
}

func main() {
	port := flag.String("p", "3333", "port to serve")
	redisURL := flag.String("u", "db-url:6379", "redis database address with port")
	redisID := flag.String("i", "db-id:6379", "redis database address with port")
	expire := flag.Int("e", 3600, "redis key timeout")
	flag.Parse()

	serverPort := os.Getenv("SERVER_PORT")
	if len(serverPort) == 0 {
		serverPort = *port
	}
	redisURLAddr := os.Getenv("REDIS_ADDR_URL")
	if len(redisURLAddr) == 0 {
		redisURLAddr = *redisURL
	}
	redisIDAddr := os.Getenv("REDIS_ADDR_ID")
	if len(redisIDAddr) == 0 {
		redisIDAddr = *redisID
	}
	redisExpire, err := strconv.Atoi(os.Getenv("REDIS_EXPIRE"))
	if redisExpire <= 0 || err != nil {
		redisExpire = *expire
	}

	poolURL, err := radix.NewPool("tcp", redisURLAddr, 10)
	if err != nil {
		log.Panic(err)
	}
	poolID, err := radix.NewPool("tcp", redisIDAddr, 10)
	if err != nil {
		log.Panic(err)
	}

	go pSaltRoutine(int(math.Ceil(float64(redisExpire) / 3)))

	ru := &redisURLStorage{
		poolID:    poolID,
		poolURL:   poolURL,
		expireURL: strconv.Itoa(redisExpire),
		expireID:  strconv.Itoa(int(math.Ceil(float64(redisExpire) / 2))),
		timeout:   int(math.Ceil(float64(redisExpire) / 3)),
	}
	http.HandleFunc("/url/", urlHandler(ru))
	http.HandleFunc("/register", registerHandler(ru))

	log.Printf("Serving on port %s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
