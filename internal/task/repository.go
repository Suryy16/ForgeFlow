package task

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTaskNotFound      = errors.New("task not found")
	ErrProjectNotFound   = errors.New("project not found")
	ErrAssigneeNotMember = errors.New("assignee is not a project member")
)

type TaskRepository interface {
	Create(ctx context.Context, task Task) (Task, error)
	GetByID(ctx context.Context, id uuid.UUID) (Task, error)
	GetAll(ctx context.Context) ([]Task, error)
	GetAllByUser(ctx context.Context, userID uuid.UUID) ([]Task, error)
	Update(ctx context.Context, task Task) (Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (Task, error) {
	var t Task

	query := `
		SELECT
			id,
			project_id,
			assignee_id,
			title,
			description,
			status,
			priority,
			due_date,
			created_at,
			updated_at
		FROM tasks
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID,
		&t.ProjectID,
		&t.AssigneeID,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.Priority,
		&t.DueDate,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrTaskNotFound
	}
	if err != nil {
		return Task{}, err
	}

	return t, nil
}

func (r *PostgresRepository) GetAll(
	ctx context.Context,
) ([]Task, error) {
	query := `
		SELECT
			id,
			project_id,
			assignee_id,
			title,
			description,
			status,
			priority,
			due_date,
			created_at,
			updated_at
		FROM tasks
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]Task, 0)

	for rows.Next() {
		var t Task

		if err := rows.Scan(
			&t.ID,
			&t.ProjectID,
			&t.AssigneeID,
			&t.Title,
			&t.Description,
			&t.Status,
			&t.Priority,
			&t.DueDate,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}

func (r *PostgresRepository) Create(
	ctx context.Context,
	task Task,
) (Task, error) {
	const query = `
		INSERT INTO tasks (
			project_id,
			assignee_id,
			title,
			description,
			status,
			priority
		)
		VALUES ($1, $2, $3, $4, 'TODO', $5)
		RETURNING
			id,
			project_id,
			assignee_id,
			title,
			description,
			status,
			priority,
			due_date,
			created_at,
			updated_at
	`

	var created Task

	err := r.db.QueryRow(
		ctx,
		query,
		task.ProjectID,
		task.AssigneeID,
		task.Title,
		task.Description,
		task.Priority,
	).Scan(
		&created.ID,
		&created.ProjectID,
		&created.AssigneeID,
		&created.Title,
		&created.Description,
		&created.Status,
		&created.Priority,
		&created.DueDate,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			switch pgErr.ConstraintName {
			case "tasks_project_id_fkey":
				return Task{}, ErrProjectNotFound
			case "fk_tasks_assignee_membership":
				return Task{}, ErrAssigneeNotMember
			}
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return Task{}, fmt.Errorf("create task returned no row: %w", err)
		}

		return Task{}, fmt.Errorf("insert task: %w", err)
	}

	return created, nil
}

func (r *PostgresRepository) Update(
	ctx context.Context,
	task Task,
) (Task, error) {
	query := `
		UPDATE tasks
		SET
			assignee_id = $2,
			title = $3,
			description = $4,
			status = $5,
			priority = $6,
			due_date = $7,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id,
			project_id,
			assignee_id,
			title,
			description,
			status,
			priority,
			due_date,
			created_at,
			updated_at
	`

	var updated Task

	err := r.db.QueryRow(
		ctx,
		query,
		task.ID,
		task.AssigneeID,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.DueDate,
	).Scan(
		&updated.ID,
		&updated.ProjectID,
		&updated.AssigneeID,
		&updated.Title,
		&updated.Description,
		&updated.Status,
		&updated.Priority,
		&updated.DueDate,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrTaskNotFound
	}
	if err != nil {
		return Task{}, fmt.Errorf("update task: %w", err)
	}

	return updated, nil
}

func (r *PostgresRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	query := `
		DELETE FROM tasks
		WHERE id = $1
	`
	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *PostgresRepository) GetAllByUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]Task, error) {
	query := `
		SELECT
			t.id,
			t.project_id,
			t.assignee_id,
			t.title,
			t.description,
			t.status,
			t.priority,
			t.due_date,
			t.created_at,
			t.updated_at
		FROM tasks t
		INNER JOIN project_members pm
			ON pm.project_id = t.project_id
		WHERE pm.user_id = $1
		ORDER BY t.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query user tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]Task, 0)

	for rows.Next() {
		var t Task

		if err := rows.Scan(
			&t.ID,
			&t.ProjectID,
			&t.AssigneeID,
			&t.Title,
			&t.Description,
			&t.Status,
			&t.Priority,
			&t.DueDate,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}
