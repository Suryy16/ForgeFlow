package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Suryy16/ForgeFlow/internal/config"
	"github.com/Suryy16/ForgeFlow/internal/database"
	"github.com/Suryy16/ForgeFlow/internal/task"
	"github.com/Suryy16/ForgeFlow/internal/user"
)

func main() {
	dsn, err := config.DatabaseURL()
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.Connect(dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := http.NewServeMux()
	taskRepository := task.NewRepository
	userRepository := user.NewRepository
	taskHandler := task.NewHandler(taskRepository())
	userHandler := user.NewHandler(userRepository())

	fmt.Println("Starting server on :8080...")
	//Tasks
	router.HandleFunc(
		"GET /api/v1/tasks", taskHandler.GetTasks,
	)
	router.HandleFunc(
		"GET /api/v1/task", taskHandler.GetTask,
	)
	router.HandleFunc(
		"POST /api/v1/task/create", taskHandler.CreateTask,
	)
	router.HandleFunc(
		"PATCH /api/v1/task/update", taskHandler.UpdateTask,
	)
	router.HandleFunc(
		"DELETE /api/v1/task/delete", taskHandler.DeleteTask,
	)

	//Users
	router.HandleFunc(
		"GET /api/v1/users", userHandler.GetAllUsers,
	)
	router.HandleFunc(
		"GET /api/v1/user", userHandler.GetUserByID,
	)
	router.HandleFunc(
		"POST /api/v1/user/create", userHandler.CreateUser,
	)

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
