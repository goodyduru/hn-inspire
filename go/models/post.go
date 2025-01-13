package models

import "time"

type Post struct {
	ID         int
	Title      string
	Url        string
	Text       string
	Votes      int
	Is_Flagged bool
	AuthorID   int
	Author     string
	CreatedAt  time.Time
}

func (p *Post) Create() error {
	_, err := db.Exec(`INSERT INTO posts (title, url, text, votes, author_id) VALUES ($1, $2, $3, $4, $5)`,
		p.Title, p.Url, p.Text, 1, p.AuthorID)
	return err
}
