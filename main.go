package main

import (
	"encoding/json"
	"io"
	"k8s/practice/repository"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

type App struct {
	db *repository.DB
}

func main() {
	db, err := repository.New()
	if err != nil {
		log.Fatalf("couldn't initiate new instance of db: %v", err)
	}
	app := &App{db: db}
	http.HandleFunc("/health", app.handleHealth)
	http.HandleFunc("/todo", app.todoHandler)
	http.HandleFunc("/listtodo", app.handleListToDo)
	http.ListenAndServe(":8080", nil)
}
func (app *App) todoHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		app.handleCreateTodo(w, r)
	case http.MethodPut:
		app.handleUpdateTodo(w, r)
	case http.MethodDelete:
		app.handleDeleteTodo(w, r)
	default:
		http.Error(w, "Method is not supperted", http.StatusMethodNotAllowed)
	}
}
func (app *App) handleCreateTodo(w http.ResponseWriter, r *http.Request) {
	var todo repository.ToDo

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	err = json.Unmarshal(body, &todo)
	if err != nil {
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := app.db.CreateToDo(todo.Name); err != nil {
		http.Error(w, "Failed to create TO DO", http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusCreated)
}
func (app *App) handleUpdateTodo(w http.ResponseWriter, r *http.Request) {
	var todo repository.ToDo

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	err = json.Unmarshal(body, &todo)
	if err != nil {
		http.Error(w, "invalid JSON format", http.StatusBadRequest)
		return
	}
	id, _ := strconv.Atoi(todo.Id)
	if err := app.db.UpdateTodo(id, todo.Name); err != nil {
		http.Error(w, "Failed to update TO DO", http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusNoContent)
}
func (app *App) handleDeleteTodo(w http.ResponseWriter, r *http.Request) {

	toDoId := r.URL.Query().Get("id")
	if toDoId == "" {
		http.Error(w, "id can't be empty", http.StatusBadRequest)
	}
	id, _ := strconv.Atoi(toDoId)
	if err := app.db.DeleteTodo(id); err != nil {
		http.Error(w, "Failed to delete TO DO", http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusNoContent)
}
func (app *App) handleListToDo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method is not supported", http.StatusInternalServerError)
	}
	todos, err := app.db.ListToDo()
	if err != nil {
		http.Error(w, "failed to get TO DO list", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	responseData, err := json.Marshal(todos)
	if err != nil {
		http.Error(w, "failed to marshal data", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
}
func (app *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
