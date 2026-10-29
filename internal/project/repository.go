package project

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
	ErrProjectNotFound = errors.New("project not found")
	ErrOwnerNotMember  = errors.New("owner is not a project member")
	ErrUserNotFound    = errors.New("user not found")
	ErrMemberExists    = errors.New("user is already a project member")
)

type ProjectRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (Project, error)
	GetAll(ctx context.Context) ([]Project, error)
	Create(ctx context.Context, project Project) (Project, error)
	Update(ctx context.Context, project Project) (Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AddMember(ctx context.Context, member ProjectMember) (ProjectMember, error)
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
) (Project, error) {
	var p Project

	query := `
		SELECT
			id,
			owner_id,
			name,
			description,
			created_at,
			updated_at
		FROM projects
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.OwnerID,
		&p.Name,
		&p.Description,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Project{}, ErrProjectNotFound
	}
	if err != nil {
		return Project{}, err
	}
	return p, nil
}

func (r *PostgresRepository) GetAll(
	ctx context.Context,
) ([]Project, error) {
	query := `
		SELECT
			id,
			owner_id,
			name,
			description,
			created_at,
			updated_at
		FROM projects
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query projects: %w", err)
	}
	defer rows.Close()

	projects := make([]Project, 0)

	for rows.Next() {
		var p Project

		if err := rows.Scan(
			&p.ID,
			&p.OwnerID,
			&p.Name,
			&p.Description,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("Scan project: %w", err)
		}
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}
	return projects, nil
}

func (r *PostgresRepository) Create(
	ctx context.Context,
	project Project,
) (Project, error) {
	const query = `
		INSERT INTO projects (
			name,
			description,
			owner_id
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			name,
			description,
			owner_id,
			created_at,
			updated_at
	`

	var created Project

	err := r.db.QueryRow(
		ctx,
		query,
		project.Name,
		project.Description,
		project.OwnerID,
	).Scan(
		&created.ID,
		&created.Name,
		&created.Description,
		&created.OwnerID,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23503" {
				switch pgErr.ConstraintName {
				case "fk_projects_owner_membership":
					return Project{}, ErrOwnerNotMember
				}
			}
		}

		return Project{}, fmt.Errorf("insert project: %w", err)
	}

	return created, nil
}

func (r *PostgresRepository) Update(
	ctx context.Context,
	project Project,
) (Project, error) {
	query := `
		UPDATE projects
		SET
			owner_id = $2,
			name = $3,
			description = $4,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id,
			name,
			description,
			owner_id,
			created_at,
			updated_at
	`

	var updated Project

	err := r.db.QueryRow(
		ctx,
		query,
		project.ID,
		project.OwnerID,
		project.Name,
		project.Description,
	).Scan(
		&updated.ID,
		&updated.Name,
		&updated.Description,
		&updated.OwnerID,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Project{}, ErrProjectNotFound
	}

	if err != nil {
		return Project{}, fmt.Errorf("update project: %w", err)
	}

	return updated, nil
}

func (r *PostgresRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	query := `
		DELETE FROM projects
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrProjectNotFound
	}
	return nil
}

func (r *PostgresRepository) AddMember(
	ctx context.Context,
	member ProjectMember,
) (ProjectMember, error) {
	const query = `
		INSERT INTO project_members (
			project_id,
			user_id,
			role
		)
		VALUES ($1, $2, $3)
		RETURNING
			project_id,
			user_id,
			role,
			created_at,
			updated_at
	`

	var created ProjectMember

	err := r.db.QueryRow(
		ctx,
		query,
		member.ProjectID,
		member.UserID,
		member.Role,
	).Scan(
		&created.ProjectID,
		&created.UserID,
		&created.Role,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch {
			case pgErr.Code == "23505" &&
				pgErr.ConstraintName == "project_members_pkey":
				return ProjectMember{}, ErrMemberExists

			case pgErr.Code == "23503":
				switch pgErr.ConstraintName {
				case "project_members_project_id_fkey":
					return ProjectMember{}, ErrProjectNotFound
				case "project_members_user_id_fkey":
					return ProjectMember{}, ErrUserNotFound
				}
			}
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return ProjectMember{}, fmt.Errorf(
				"insert project member returned no row: %w",
				err,
			)
		}

		return ProjectMember{}, fmt.Errorf("insert project member: %w", err)
	}

	return created, nil
}
