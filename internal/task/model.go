package task

import "time"

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
