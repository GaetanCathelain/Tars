package db

import (
	"context"
	"fmt"
	"time"

	"github.com/GaetanCathelain/Tars/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Queries struct {
	pool *pgxpool.Pool
}

func NewQueries(pool *pgxpool.Pool) *Queries {
	return &Queries{pool: pool}
}

// --- Users ---

func (q *Queries) UpsertUser(ctx context.Context, u *models.User) error {
	return q.pool.QueryRow(ctx, `
		INSERT INTO users (github_id, username, email, avatar_url, access_token)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (github_id) DO UPDATE SET
			username = EXCLUDED.username,
			email = EXCLUDED.email,
			avatar_url = EXCLUDED.avatar_url,
			access_token = EXCLUDED.access_token,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`, u.GitHubID, u.Username, u.Email, u.AvatarURL, u.AccessToken,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	u := &models.User{}
	err := q.pool.QueryRow(ctx, `
		SELECT id, github_id, username, email, avatar_url, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.GitHubID, &u.Username, &u.Email, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return u, err
}

// --- Sessions ---

func (q *Queries) CreateSession(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (*models.Session, error) {
	s := &models.Session{}
	err := q.pool.QueryRow(ctx, `
		INSERT INTO sessions (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token, expires_at, created_at
	`, userID, token, expiresAt).Scan(&s.ID, &s.UserID, &s.Token, &s.ExpiresAt, &s.CreatedAt)
	return s, err
}

func (q *Queries) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	s := &models.Session{}
	err := q.pool.QueryRow(ctx, `
		SELECT id, user_id, token, expires_at, created_at
		FROM sessions WHERE token = $1 AND expires_at > NOW()
	`, token).Scan(&s.ID, &s.UserID, &s.Token, &s.ExpiresAt, &s.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (q *Queries) DeleteSession(ctx context.Context, token string) error {
	_, err := q.pool.Exec(ctx, "DELETE FROM sessions WHERE token = $1", token)
	return err
}

// --- Repos ---

func (q *Queries) ListRepos(ctx context.Context) ([]models.Repo, error) {
	rows, err := q.pool.Query(ctx, `
		SELECT id, name, url, local_path, default_branch, added_by, created_at
		FROM repos ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []models.Repo
	for rows.Next() {
		var r models.Repo
		if err := rows.Scan(&r.ID, &r.Name, &r.URL, &r.LocalPath, &r.DefaultBranch, &r.AddedBy, &r.CreatedAt); err != nil {
			return nil, err
		}
		repos = append(repos, r)
	}
	return repos, rows.Err()
}

func (q *Queries) CreateRepo(ctx context.Context, r *models.Repo) error {
	return q.pool.QueryRow(ctx, `
		INSERT INTO repos (name, url, local_path, default_branch, added_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, r.Name, r.URL, r.LocalPath, r.DefaultBranch, r.AddedBy).Scan(&r.ID, &r.CreatedAt)
}

func (q *Queries) GetRepoByID(ctx context.Context, id uuid.UUID) (*models.Repo, error) {
	r := &models.Repo{}
	err := q.pool.QueryRow(ctx, `
		SELECT id, name, url, local_path, default_branch, added_by, created_at
		FROM repos WHERE id = $1
	`, id).Scan(&r.ID, &r.Name, &r.URL, &r.LocalPath, &r.DefaultBranch, &r.AddedBy, &r.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return r, err
}

func (q *Queries) DeleteRepo(ctx context.Context, id uuid.UUID) error {
	tag, err := q.pool.Exec(ctx, "DELETE FROM repos WHERE id = $1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("repo not found")
	}
	return nil
}

// --- Tasks ---

func (q *Queries) ListTasks(ctx context.Context, repoID *uuid.UUID, status *string) ([]models.Task, error) {
	query := `SELECT id, repo_id, title, description, status, priority, created_by, created_at, updated_at FROM tasks WHERE 1=1`
	args := []any{}
	argN := 1

	if repoID != nil {
		query += fmt.Sprintf(" AND repo_id = $%d", argN)
		args = append(args, *repoID)
		argN++
	}
	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argN)
		args = append(args, *status)
		argN++
	}
	query += " ORDER BY priority DESC, created_at DESC"

	rows, err := q.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.RepoID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (q *Queries) CreateTask(ctx context.Context, t *models.Task) error {
	return q.pool.QueryRow(ctx, `
		INSERT INTO tasks (repo_id, title, description, status, priority, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`, t.RepoID, t.Title, t.Description, t.Status, t.Priority, t.CreatedBy).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (q *Queries) GetTaskByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	t := &models.Task{}
	err := q.pool.QueryRow(ctx, `
		SELECT id, repo_id, title, description, status, priority, created_by, created_at, updated_at
		FROM tasks WHERE id = $1
	`, id).Scan(&t.ID, &t.RepoID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func (q *Queries) UpdateTask(ctx context.Context, id uuid.UUID, status *string, title *string, description *string, priority *int) (*models.Task, error) {
	// Build dynamic update.
	sets := []string{}
	args := []any{}
	argN := 1

	if status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", argN))
		args = append(args, *status)
		argN++
	}
	if title != nil {
		sets = append(sets, fmt.Sprintf("title = $%d", argN))
		args = append(args, *title)
		argN++
	}
	if description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", argN))
		args = append(args, *description)
		argN++
	}
	if priority != nil {
		sets = append(sets, fmt.Sprintf("priority = $%d", argN))
		args = append(args, *priority)
		argN++
	}

	if len(sets) == 0 {
		return q.GetTaskByID(ctx, id)
	}

	sets = append(sets, "updated_at = NOW()")
	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = $%d RETURNING id, repo_id, title, description, status, priority, created_by, created_at, updated_at",
		joinStrings(sets, ", "), argN)
	args = append(args, id)

	t := &models.Task{}
	err := q.pool.QueryRow(ctx, query, args...).Scan(
		&t.ID, &t.RepoID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return t, err
}

func joinStrings(s []string, sep string) string {
	result := ""
	for i, v := range s {
		if i > 0 {
			result += sep
		}
		result += v
	}
	return result
}
