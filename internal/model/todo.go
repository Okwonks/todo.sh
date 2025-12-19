package model

import "time"

type status string

const (
	Started status = "started"
	Blocked status = "blocked"
	Done    status = "done"
)

type Todo struct {
	ID          int64      `json:"id"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	Description string     `json:"description"`
	Completed   bool       `json:"completed"`
  DueDate     *time.Time `json:"due_date,omitempty"`
	Priority    int        `json:"priority"`
	Status      status     `json:"status"` 
}
