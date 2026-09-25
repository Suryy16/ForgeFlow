package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Suryy16/ForgeFlow/internal/task"
)

func main() {
	taskRepository := task.NewRepository
	taskHandler := task.NewHandler(taskRepository())

	fmt.Println("Starting server on :8080...")
	http.HandleFunc("/api/v1/tasks", taskHandler.GetTasks)
	http.HandleFunc("/api/v1/task", taskHandler.GetTask)
	http.HandleFunc("/api/v1/task/create", taskHandler.CreateTask)
	http.HandleFunc("/api/v1/task/update", taskHandler.UpdateTask)
	http.HandleFunc("/api/v1/task/delete", taskHandler.DeleteTask)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
