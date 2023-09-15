package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/adamaho/realtime/pkg/realtime"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type data struct {
	Count int `json:"count"`
}

const SESSION_ID = "count"

func main() {
	r := chi.NewRouter()

	c := data{Count: 1}
	rt := realtime.New()

	r.Use(middleware.Logger)

	j, _ := json.Marshal(c)

	r.Get("/count", func(w http.ResponseWriter, r *http.Request) {
		rt.Response(w, r, j, SESSION_ID, realtime.ResponseOptions(realtime.WithBufferSize(0)))
	})

	r.Post("/count/foo", func(w http.ResponseWriter, r *http.Request) {
		c.Count += 1
		fmt.Println(c.Count)
		w.Write([]byte("count updated."))
	})

	r.Post("/count/increment", func(w http.ResponseWriter, r *http.Request) {
		c.Count += 1
		j, _ := json.Marshal(c)
		rt.SendMessage(j, SESSION_ID)
		w.Write([]byte("count updated."))
	})

	r.Post("/count/decrement", func(w http.ResponseWriter, r *http.Request) {
		c.Count -= 1
		j, _ := json.Marshal(c)
		rt.SendMessage(j, SESSION_ID)
		w.Write([]byte("count updated."))
	})

	fmt.Println("Server started at https://localhost:3000")
	err := http.ListenAndServeTLS(":3000", "cert.pem", "key.pem", r)

	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
