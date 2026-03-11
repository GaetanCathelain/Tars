package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Task struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedBy uuid.UUID `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID         uuid.UUID  `json:"id"`
	TaskID     uuid.UUID  `json:"task_id"`
	SenderType string     `json:"sender_type"`
	SenderID   *uuid.UUID `json:"sender_id,omitempty"`
	Content    string     `json:"content"`
	CreatedAt  time.Time  `json:"created_at"`
}

type WorkerSession struct {
	ID         uuid.UUID  `json:"id"`
	TaskID     uuid.UUID  `json:"task_id"`
	MessageID  *uuid.UUID `json:"message_id,omitempty"`
	Status     string     `json:"status"`
	Command    string     `json:"command"`
	ExitCode   *int       `json:"exit_code,omitempty"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

type WorkerOutput struct {
	ID        int64     `json:"id"`
	SessionID uuid.UUID `json:"session_id"`
	Data      []byte    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
}
