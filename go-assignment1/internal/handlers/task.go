package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

var tasks = map[int]Task{
	1:  {ID: 1, Title: "Write unit tests", Done: false},
	2:  {ID: 2, Title: "Deploy service", Done: true},
	42: {ID: 42, Title: "Write unit tests", Done: false},
}

func TasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTasks(w, r)
	case http.MethodPost:
		createTask(w, r)
	case http.MethodPatch:
		updateTask(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}

}

func getTasks(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")

	if idStr == "" {
		// no id: return all tasks
		allTasks := tasks
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(allTasks)
		return
	}

	// parse id to integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// invalid id, not int
		sendError(w, http.StatusBadRequest, "invalid id")
		return
	}

	task, exists := tasks[id]
	if !exists {
		sendError(w, http.StatusNotFound, "task not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.Title == "" {
		sendError(w, http.StatusBadRequest, "title is required")
		return
	}
	// Create new ID
	nextID := 1
	for id := range tasks {
		if id >= nextID {
			nextID = id + 1
		}
	}
	// Create new task
	newTask := Task{
		ID:    nextID,
		Title: req.Title,
		Done:  false,
	}
	tasks[nextID] = newTask

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTask)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")

	if idStr == "" {
		sendError(w, http.StatusBadRequest, "invalid id")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req struct {
		Done bool `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid json")
		return
	}

	task, exists := tasks[id]
	if !exists {
		sendError(w, http.StatusNotFound, "task not found")
		return
	}
	// update done status
	task.Done = req.Done
	tasks[id] = task

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
