package tweet

import "time"

type Tweet struct {
	ID        string
	UserID    string
	Content   string
	CreatedAt time.Time
}

type CreateInput struct {
	UserID  string
	Content string
}
