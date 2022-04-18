package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/handlers"
)

const (
	MethodGet    = "GET"
	MethodPost   = "POST"
	MethodPut    = "PUT"
	MethodDelete = "DELETE"
)

type Resource interface {
	Get(values url.Values, headers http.Header) (int, interface{})
	Post(values url.Values) (int, interface{})
	Put(values url.Values) (int, interface{})
	Delete(values url.Values) (int, interface{})
}

type ResourceBase struct{}

func (ResourceBase) Get(values url.Values) (int, interface{}) {
	return http.StatusMethodNotAllowed, ""
}

func (ResourceBase) Post(values url.Values) (int, interface{}) {
	return http.StatusMethodNotAllowed, ""
}

func (ResourceBase) Put(values url.Values) (int, interface{}) {
	return http.StatusMethodNotAllowed, ""
}

func (ResourceBase) Delete(values url.Values) (int, interface{}) {
	return http.StatusMethodNotAllowed, ""
}

func requestHandler(resource Resource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var data interface{}
		var status int

		r.ParseForm()
		method := r.Method
		values := r.Form
		headers := r.Header

		switch method {
		case MethodGet:
			status, data = resource.Get(values, headers)
		case MethodPost:
			status, data = resource.Post(values)
		case MethodPut:
			status, data = resource.Put(values)
		case MethodDelete:
			status, data = resource.Delete(values)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		content, err := json.Marshal(data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)
		w.Write(content)
	}
}

func AddResource(resource Resource, path string) {
	http.HandleFunc(path, requestHandler(resource))
}

func StartServer(port int) {
	portString := fmt.Sprintf(":%d", port)
	err := http.ListenAndServe(portString, handlers.LoggingHandler(os.Stdout, http.DefaultServeMux))
	if err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Auth resource
type Auth struct {
	ResourceBase
}

// Auth GET method
func (a Auth) Get(values url.Values, headers http.Header) (int, interface{}) {
	var r interface{}
	var status int
	var authed bool
	var cookie string

	redis_host := getEnv("REDIS_HOST", "localhost")
	redis_port := getEnv("REDIS_PORT", "6379")

	cookie_header := headers.Get("cookie")
	if cookie_header == "" || !strings.Contains(cookie_header, "wordpress_logged_in_") {
		status = 401
		return status, r
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redis_host + ":" + redis_port,
		Password: "",
		DB:       0,
	})

	iter := rdb.Scan(ctx, 0, "wp-cookie-*", 0).Iterator()
	for iter.Next(ctx) {
		//log.Println("keys", iter.Val())
		val, err := rdb.Get(ctx, iter.Val()).Result()
		if err != nil {
			panic(err)
		}
		//log.Println("cookie", cookie)
		//log.Println("key   ", "wordpress_logged_in_"+val)
		cookie_slice := strings.Split(cookie_header, ";")
		for _, value := range cookie_slice {
			//log.Println(value)
			if strings.Contains(value, "wordpress_logged_in_") {
				cookie = strings.TrimSpace(value)
				break
			} else {
				cookie = ""
			}
		}
		if cookie == "wordpress_logged_in_"+val {
			authed = true
			log.Println("Success, cookie matched: " + cookie)
			break
		} else {
			authed = false
		}
	}

	if err := iter.Err(); err != nil {
		panic(err)
	} else if authed == true {
		status = 200
	} else {
		status = 401
		log.Println("Failure, cookie not found in redis: " + cookie)
	}

	return status, r
}

var ctx = context.Background()

func main() {
	var auth Auth
	AddResource(auth, "/auth")
	StartServer(8081)
}
