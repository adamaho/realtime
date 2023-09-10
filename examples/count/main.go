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

func updateCount(w http.ResponseWriter, rt *realtime.Realtime, d *data) {
	countJson, _ := json.Marshal(d)
	rt.SendMessage(countJson, SESSION_ID)
	w.Write([]byte("count updated."))
}

func main() {
	r := chi.NewRouter()

	c := data{Count: 1}
	rt := realtime.New()

	r.Use(middleware.Logger)

	r.Get("/count", func(w http.ResponseWriter, r *http.Request) {
		json, _ := json.Marshal(c)
		rt.Response(w, r, json, SESSION_ID, realtime.ResponseOptions())
	})

	r.Post("/count/increment", func(w http.ResponseWriter, r *http.Request) {
		c.Count += 1
		updateCount(w, &rt, &c)
	})

	r.Post("/count/decrement", func(w http.ResponseWriter, r *http.Request) {
		c.Count -= 1
		updateCount(w, &rt, &c)
	})

	fmt.Println("Server started at https://localhost:3000")
	err := http.ListenAndServeTLS(":3000", "cert.pem", "key.pem", r)

	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
