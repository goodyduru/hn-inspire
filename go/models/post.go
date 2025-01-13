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
	NumReplies int
	VotedID    int
	FlaggedID  int
}

var LIMIT = 30

func (p *Post) Create() error {
	_, err := db.Exec(`INSERT INTO posts (title, url, text, votes, author_id) VALUES ($1, $2, $3, $4, $5)`,
		p.Title, p.Url, p.Text, 1, p.AuthorID)
	return err
}

func GetAuthRanked(userId int) ([]Post, error) {
	var posts []Post
	rows, err := db.Query(`WITH RECURSIVE rankedposts AS (
								SELECT posts.id, title, url, votes, created_at, username, 
								(votes - 1)/pow(((extract(epoch from (now() - created_at))/3600) + 2), 1.5) as rank
								FROM posts JOIN users ON posts.author_id=users.id WHERE parent_id IS NULL ORDER BY rank, created_at DESC LIMIT $2
							),
							comments AS (
								SELECT posts.id, rankedposts.id AS ancestor FROM posts JOIN rankedposts ON posts.parent_id=rankedposts.id
								UNION ALL
								SELECT posts.id, ancestor FROM posts JOIN comments ON posts.parent_id=comments.id
							),
							comments_count AS (
								SELECT ancestor, count(*) AS comment_count FROM comments GROUP BY ancestor
							),
							user_votes AS (
								SELECT user_id, post_id FROM votes WHERE user_id=$1
							),
							user_flags AS (
								SELECT user_id, post_id FROM flags WHERE user_id=$1
							)
							SELECT rankedposts.id, rankedposts.title, rankedposts.url, rankedposts.votes, rankedposts.created_at,
									rankedposts.username, COALESCE(comment_count, 0), COALESCE(user_votes.user_id, 0) AS user_voted, 
									COALESCE(user_flags.user_id, 0) AS user_flagged
							FROM rankedposts
							LEFT JOIN comments_count ON rankedposts.id=comments_count.ancestor
							LEFT JOIN user_votes ON rankedposts.id=user_votes.post_id
							LEFT JOIN user_flags ON rankedposts.id=user_flags.post_id
							`, userId, LIMIT)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Url, &post.Votes, &post.CreatedAt, &post.Author,
			&post.NumReplies, &post.VotedID, &post.FlaggedID); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func GetRanked() ([]Post, error) {
	return nil, nil
}
