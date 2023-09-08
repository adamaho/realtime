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

func updateCount(w http.ResponseWriter, r *http.Request, rt *realtime.Realtime, d data) {
	var msg json.RawMessage

	p := r.URL.Query().Get("patch")
	if p == "true" {
		f, _ := rt.CreatePatch(d, SESSION_ID)
		msg = f
	} else {
		countJson, _ := json.Marshal(d)
		msg = countJson
	}

	rt.PublishMsg(msg, SESSION_ID)
	w.Write([]byte("count updated."))
}

func main() {
	r := chi.NewRouter()

	c := data{Count: 1}
	rt := realtime.New()

	r.Use(middleware.Logger)

	r.Get("/count", func(w http.ResponseWriter, r *http.Request) {
		rt.Stream(w, r, c, SESSION_ID, true)
	})

	r.Post("/count/increment", func(w http.ResponseWriter, r *http.Request) {
		c.Count += 1
		updateCount(w, r, &rt, c)
	})

	r.Post("/count/decrement", func(w http.ResponseWriter, r *http.Request) {
		c.Count -= 1
		updateCount(w, r, &rt, c)
	})

	fmt.Println("Server started at https://localhost:3000")
	err := http.ListenAndServeTLS(":3000", "cert.pem", "key.pem", r)

	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
