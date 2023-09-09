package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/adamaho/realtime/pkg/realtime"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
)

type todo struct {
	Id          uuid.UUID `json:"todo_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Checked     bool      `json:"checked"`
}

const SESSION_ID = "todos"

func mutationResponse(w http.ResponseWriter, rt *realtime.Realtime, d *[]todo) {
	countJson, _ := json.Marshal(d)

	f, _ := rt.CreatePatch(countJson, SESSION_ID)
	if len(f) != 0 {
		rt.SendMessage(f, SESSION_ID)
	}

	w.Write([]byte("todos updated."))
}

func main() {
	r := chi.NewRouter()

	todos := make([]todo, 0)
	rt := realtime.New()

	r.Use(middleware.Logger)

	r.Get("/todos", func(w http.ResponseWriter, r *http.Request) {
		json, _ := json.Marshal(todos)
		rt.Stream(w, r, json, SESSION_ID, true)
	})

	r.Post("/todos", func(w http.ResponseWriter, r *http.Request) {
		var newTodo todo

		err := json.NewDecoder(r.Body).Decode(&newTodo)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		todo := todo{
			Id:          uuid.New(),
			Title:       newTodo.Title,
			Description: newTodo.Description,
			Checked:     false,
		}

		todos = append(todos, todo)

		mutationResponse(w, &rt, &todos)
	})

	r.Put("/todos/{todoID}", func(w http.ResponseWriter, r *http.Request) {
		todoIDParam := chi.URLParam(r, "todoID")
		todoID, _ := uuid.Parse(todoIDParam)

		var todoUpdate todo

		err := json.NewDecoder(r.Body).Decode(&todoUpdate)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		for index, todo := range todos {
			if todo.Id == todoID {
				todos[index].Title = todoUpdate.Title
				todos[index].Description = todoUpdate.Description
				todos[index].Checked = todoUpdate.Checked
				mutationResponse(w, &rt, &todos)
				return
			}
		}

		http.Error(w, "Todo not found", http.StatusNotFound)
	})

	r.Delete("/todos/{todoID}", func(w http.ResponseWriter, r *http.Request) {
		todoIDParam := chi.URLParam(r, "todoID")
		todoID, _ := uuid.Parse(todoIDParam)

		var todoIndex int = -1
		for index, todo := range todos {
			if todo.Id == todoID {
				todoIndex = index
				break
			}
		}

		if todoIndex == -1 {
			http.Error(w, "Todo not found", http.StatusNotFound)
		}

		todos = append(todos[:todoIndex], todos[todoIndex+1:]...)

		mutationResponse(w, &rt, &todos)
	})

	fmt.Println("Server started at https://localhost:3000")
	err := http.ListenAndServeTLS(":3000", "cert.pem", "key.pem", r)

	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
