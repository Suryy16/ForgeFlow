package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Task struct {
	ID          int       `json:"id"`
	Owner_ID    int       `json:"owner_id"`
	Assignee_ID int       `json:"assignee_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	Due_date    time.Time `json:"due_date"`
	Created_at  time.Time `json:"created_at"`
	Updated_at  time.Time `json:"updated_at"`
}

var location = time.FixedZone("UTC+8", 8*60*60)

// mock data
var tasks = []Task{
	{
		ID:          1,
		Owner_ID:    101,
		Assignee_ID: 201,
		Title:       "Implementasi Login API",
		Description: "Membuat endpoint login dengan JWT authentication",
		Status:      "IN_PROGRES",
		Priority:    "HIGH",
		Due_date:    time.Date(2024, 7, 15, 17, 0, 0, 0, location),
		Created_at:  time.Date(2024, 7, 1, 9, 30, 0, 0, location),
		Updated_at:  time.Date(2024, 7, 5, 14, 20, 0, 0, location),
	},
	{
		ID:          2,
		Owner_ID:    101,
		Assignee_ID: 202,
		Title:       "Desain Halaman Dashboard",
		Description: "Membuat mockup dan UI dashboard admin menggunakan Figma",
		Status:      "TODO",
		Priority:    "MEDIUM",
		Due_date:    time.Date(2024, 7, 20, 17, 0, 0, 0, location),
		Created_at:  time.Date(2024, 7, 2, 10, 0, 0, 0, location),
		Updated_at:  time.Date(2024, 7, 2, 10, 0, 0, 0, location),
	},
	{
		ID:          3,
		Owner_ID:    102,
		Assignee_ID: 203,
		Title:       "Fix Bug pada Fitur Upload",
		Description: "File upload gagal ketika ukuran lebih dari 5MB",
		Status:      "DONE",
		Priority:    "HIGH",
		Due_date:    time.Date(2024, 7, 10, 12, 0, 0, 0, location),
		Created_at:  time.Date(2024, 6, 28, 8, 15, 0, 0, location),
		Updated_at:  time.Date(2024, 7, 9, 16, 45, 0, 0, location),
	},
	{
		ID:          4,
		Owner_ID:    102,
		Assignee_ID: 201,
		Title:       "Menulis Dokumentasi API",
		Description: "Membuat dokumentasi endpoint menggunakan Swagger",
		Status:      "TODO",
		Priority:    "LOW",
		Due_date:    time.Date(2024, 7, 30, 17, 0, 0, 0, location),
		Created_at:  time.Date(2024, 7, 3, 13, 0, 0, 0, location),
		Updated_at:  time.Date(2024, 7, 3, 13, 0, 0, 0, location),
	},
	{
		ID:          5,
		Owner_ID:    103,
		Assignee_ID: 204,
		Title:       "Setup CI/CD Pipeline",
		Description: "Konfigurasi GitHub Actions untuk deployment otomatis ke server staging",
		Status:      "IN_PROGRES",
		Priority:    "MEDIUM",
		Due_date:    time.Date(2024, 7, 25, 17, 0, 0, 0, location),
		Created_at:  time.Date(2024, 7, 4, 11, 30, 0, 0, location),
		Updated_at:  time.Date(2024, 7, 6, 9, 15, 0, 0, location),
	},
}

func main() {
	fmt.Println("Starting server on :8080...")
	http.HandleFunc("/api/v1/tasks", getTasks)
	http.HandleFunc("/api/v1/task", getTask)
	http.HandleFunc("/api/v1/task/create", createTask)
	http.HandleFunc("/api/v1/task/update", updateTask)
	http.HandleFunc("/api/v1/task/delete", deleteTask)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	for _, task := range tasks {
		if task.ID == id {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(task)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}
	http.Error(w, "task not found", http.StatusNotFound)
}

func createTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	task.ID = len(tasks) + 1
	tasks = append(tasks, task)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	var updatedtask Task
	err := json.NewDecoder(r.Body).Decode(&updatedtask)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}
	for i, task := range tasks {
		if task.ID == id {
			tasks[i].Title = updatedtask.Title
			tasks[i].Description = updatedtask.Description
			tasks[i].Status = updatedtask.Status
			tasks[i].Priority = updatedtask.Priority
			tasks[i].Due_date = updatedtask.Due_date
			tasks[i].Updated_at = time.Now().In(location).Truncate(time.Second)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(tasks[i])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}
	http.Error(w, "task not found", http.StatusNotFound)

}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	for i, task := range tasks {
		if task.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, "task not found", http.StatusNotFound)
}
