package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Suryy16/ForgeFlow/internal/auth"
	"github.com/Suryy16/ForgeFlow/internal/authorization"
	"github.com/Suryy16/ForgeFlow/internal/config"
	"github.com/Suryy16/ForgeFlow/internal/database"
	"github.com/Suryy16/ForgeFlow/internal/project"
	"github.com/Suryy16/ForgeFlow/internal/task"
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET must be set and contain at least 32 characters")
	}

	tokenTTL := 15 * time.Minute
	if raw := os.Getenv("JWT_EXPIRES_IN"); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil || parsed <= 0 {
			log.Fatal("JWT_EXPIRES_IN must be a positive duration, e.g. 15m")
		}
		tokenTTL = parsed
	}

	roleStore := authorization.NewPostgresProjectRoleStore(db)

	authRepository := auth.NewRepository(db)
	authHandler := auth.NewHandler(authRepository, jwtSecret, tokenTTL)

	taskRepository := task.NewPostgresRepository(db)
	taskHandler := task.NewHandler(taskRepository)

	projectRepository := project.NewPostgresRepository(db)
	projectHandler := project.NewHandler(projectRepository)

	protected := func(
		handler http.Handler,
		projectIDParam string,
		roles ...string,
	) http.Handler {
		roleCheck := authorization.RequireProjectRole(
			roleStore,
			projectIDParam,
			roles...,
		)(handler)

		return authHandler.RequireAuth(roleCheck)
	}

	fmt.Println("Starting server on :8080...")
	//Tasks
	router.HandleFunc(
		"GET /api/v1/tasks", taskHandler.GetTasks,
	)
	router.HandleFunc(
		"GET /api/v1/tasks/{id}", taskHandler.GetTask,
	)
	router.HandleFunc(
		"POST /api/v1/projects/{project_id}/tasks", taskHandler.CreateTask,
	)
	router.HandleFunc(
		"PUT /api/v1/tasks/{id}", taskHandler.UpdateTask,
	)
	router.HandleFunc(
		"DELETE /api/v1/tasks/{id}", taskHandler.DeleteTask,
	)

	//Projects
	router.Handle(
		"GET /api/v1/projects/{id}",
		protected(
			http.HandlerFunc(projectHandler.GetProject),
			"id",
			"OWNER", "ADMIN", "MEMBER",
		),
	)
	router.HandleFunc(
		"GET /api/v1/projects", projectHandler.GetProjects,
	)
	router.HandleFunc(
		"POST /api/v1/projects", projectHandler.CreateProject,
	)
	router.Handle(
		"PUT /api/v1/projects/{id}",
		protected(
			http.HandlerFunc(projectHandler.UpdateProject),
			"id",
			"OWNER", "ADMIN",
		),
	)
	router.Handle(
		"DELETE /api/v1/projects/{id}",
		protected(
			http.HandlerFunc(projectHandler.DeleteProject),
			"id",
			"OWNER",
		),
	)
	router.Handle(
		"POST /api/v1/projects/{project_id}/members",
		protected(
			http.HandlerFunc(projectHandler.AddMember),
			"project_id",
			"OWNER", "ADMIN",
		),
	)

	//Users
	router.HandleFunc(
		"POST /api/v1/auth/register",
		authHandler.Register,
	)
	router.HandleFunc(
		"POST /api/v1/auth/login",
		authHandler.Login,
	)
	router.Handle(
		"GET /api/v1/auth/me",
		authHandler.RequireAuth(http.HandlerFunc(authHandler.Me)),
	)

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
