package models

import "time"

type Post struct {
	ID         int
	Title      string
	Url        string
	Text       string
	Votes      int
	Is_Flagged bool
	ParentID   int
	AuthorID   int
	Author     string
	CreatedAt  time.Time
	NumReplies int
	VotedID    int
	FlaggedID  int
	HtmlClass  string
	Lev        int
}

var LIMIT = 30

func (p *Post) Create() error {
	var err error
	if p.ParentID > 0 {
		err = db.QueryRow(`INSERT INTO posts (title, url, text, votes, author_id, parent_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
			p.Title, p.Url, p.Text, 1, p.AuthorID, p.ParentID).Scan(&p.ID)
	} else {
		_, err = db.Exec(`INSERT INTO posts (title, url, text, votes, author_id) VALUES ($1, $2, $3, $4, $5)`,
			p.Title, p.Url, p.Text, 1, p.AuthorID)
	}
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
	var posts []Post
	rows, err := db.Query(`WITH RECURSIVE rankedposts AS (
								SELECT posts.id, title, url, votes, created_at, username, 
								(votes - 1)/pow(((extract(epoch from (now() - created_at))/3600) + 2), 1.5) as rank
								FROM posts JOIN users ON posts.author_id=users.id WHERE parent_id IS NULL ORDER BY rank, created_at DESC LIMIT $1
							),
							comments AS (
								SELECT posts.id, rankedposts.id AS ancestor FROM posts JOIN rankedposts ON posts.parent_id=rankedposts.id
								UNION ALL
								SELECT posts.id, ancestor FROM posts JOIN comments ON posts.parent_id=comments.id
							),
							comments_count AS (
								SELECT ancestor, count(*) AS comment_count FROM comments GROUP BY ancestor
							)
							SELECT rankedposts.id, rankedposts.title, rankedposts.url, rankedposts.votes, rankedposts.created_at,
									rankedposts.username, COALESCE(comment_count, 0) FROM rankedposts
							LEFT JOIN comments_count ON rankedposts.id=comments_count.ancestor
							`, LIMIT)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Url, &post.Votes, &post.CreatedAt, &post.Author,
			&post.NumReplies); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func GetAuthSubmitted(authorId, userId int) ([]Post, error) {
	var posts []Post
	rows, err := db.Query(`WITH RECURSIVE submitted_posts AS (
								SELECT posts.id, title, url, votes, created_at
								FROM posts WHERE author_id=$1 AND parent_id IS NULL ORDER BY created_at DESC LIMIT $3
							),
							comments AS (
								SELECT posts.id, submitted_posts.id AS ancestor FROM posts JOIN submitted_posts ON posts.parent_id=submitted_posts.id
								UNION ALL
								SELECT posts.id, ancestor FROM posts JOIN comments ON posts.parent_id=comments.id
							),
							comments_count AS (
								SELECT ancestor, count(*) AS comment_count FROM comments GROUP BY ancestor
							),
							user_votes AS (
								SELECT user_id, post_id FROM votes WHERE user_id=$2
							),
							user_flags AS (
								SELECT user_id, post_id FROM flags WHERE user_id=$2
							)
							SELECT submitted_posts.id, submitted_posts.title, submitted_posts.url, submitted_posts.votes, submitted_posts.created_at, 
								COALESCE(comment_count, 0), COALESCE(user_votes.user_id, 0), COALESCE(user_flags.user_id, 0)
							FROM submitted_posts
							LEFT JOIN comments_count ON submitted_posts.id=comments_count.ancestor
							LEFT JOIN user_votes ON submitted_posts.id=user_votes.post_id
							LEFT JOIN user_flags ON submitted_posts.id=user_flags.post_id
							`, authorId, userId, LIMIT)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Url, &post.Votes, &post.CreatedAt,
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

func GetSubmitted(authorId int) ([]Post, error) {
	var posts []Post
	rows, err := db.Query(`WITH RECURSIVE submitted_posts AS (
								SELECT posts.id, title, url, votes, created_at
								FROM posts WHERE author_id=$1 AND parent_id IS NULL ORDER BY created_at DESC LIMIT $2
							),
							comments AS (
								SELECT posts.id, submitted_posts.id AS ancestor FROM posts JOIN submitted_posts ON posts.parent_id=submitted_posts.id
								UNION ALL
								SELECT posts.id, ancestor FROM posts JOIN comments ON posts.parent_id=comments.id
							),
							comments_count AS (
								SELECT ancestor, count(*) AS comment_count FROM comments GROUP BY ancestor
							)
							SELECT submitted_posts.id, submitted_posts.title, submitted_posts.url, submitted_posts.votes, submitted_posts.created_at, 
								COALESCE(comment_count, 0) FROM submitted_posts
							LEFT JOIN comments_count ON submitted_posts.id=comments_count.ancestor
							`, authorId, LIMIT)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Url, &post.Votes, &post.CreatedAt,
			&post.NumReplies); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func GetAuthMixed(userId, currentUserId int) ([]Post, error) {
	var posts []Post
	rows, err := db.Query(`WITH RECURSIVE favorited_posts AS (
								SELECT posts.id, title, url, text, votes, author_id, parent_id, created_at
								FROM votes JOIN posts ON votes.post_id=posts.id WHERE user_id=$1 
								AND amount > 0 ORDER BY created_at DESC LIMIT $3
							),
							comments AS (
								SELECT posts.id, favorited_posts.id AS ancestor FROM posts JOIN favorited_posts ON posts.parent_id=favorited_posts.id
								UNION ALL
								SELECT posts.id, ancestor FROM posts JOIN comments ON posts.parent_id=comments.id
							),
							comments_count AS (
								SELECT ancestor, count(*) AS comment_count FROM comments GROUP BY ancestor
							),
							user_votes AS (
								SELECT user_id, post_id FROM votes WHERE user_id=$2
							),
							user_flags AS (
								SELECT user_id, post_id FROM flags WHERE user_id=$2
							)
							SELECT favorited_posts.id, favorited_posts.title, favorited_posts.url, favorited_posts.text,
							 	favorited_posts.votes, COALESCE(favorited_posts.parent_id, 0), favorited_posts.created_at, 
								COALESCE(comment_count, 0), COALESCE(user_votes.user_id, 0), users.username, 
								COALESCE(user_flags.user_id, 0)
							FROM favorited_posts
							LEFT JOIN comments_count ON favorited_posts.id=comments_count.ancestor
							LEFT JOIN user_votes ON favorited_posts.id=user_votes.post_id
							LEFT JOIN users ON favorited_posts.author_id=users.id
							LEFT JOIN user_flags ON favorited_posts.id=user_flags.post_id
							`, userId, currentUserId, LIMIT)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Url, &post.Text, &post.Votes, &post.ParentID,
			&post.CreatedAt, &post.NumReplies, &post.VotedID, &post.AuthorID, &post.FlaggedID); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func GetMixed(userId int) ([]Post, error) {
	var posts []Post
	rows, err := db.Query(`WITH RECURSIVE favorited_posts AS (
								SELECT posts.id, title, url, text, votes, author_id, parent_id, created_at
								FROM votes JOIN posts ON votes.post_id=posts.id WHERE user_id=$1 
								AND amount > 0 ORDER BY created_at DESC LIMIT $2
							),
							comments AS (
								SELECT posts.id, favorited_posts.id AS ancestor FROM posts JOIN favorited_posts ON posts.parent_id=favorited_posts.id
								UNION ALL
								SELECT posts.id, ancestor FROM posts JOIN comments ON posts.parent_id=comments.id
							),
							comments_count AS (
								SELECT ancestor, count(*) AS comment_count FROM comments GROUP BY ancestor
							)
							SELECT favorited_posts.id, favorited_posts.title, favorited_posts.url, favorited_posts.text,
							 	favorited_posts.votes, COALESCE(favorited_posts.parent_id, 0), favorited_posts.created_at, 
								COALESCE(comment_count, 0), users.username FROM favorited_posts
							LEFT JOIN comments_count ON favorited_posts.id=comments_count.ancestor
							LEFT JOIN users ON favorited_posts.author_id=users.id
							`, userId, LIMIT)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Url, &post.Text, &post.Votes, &post.ParentID,
			&post.CreatedAt, &post.NumReplies, &post.AuthorID); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func GetAuthSubmittedComments(authorId, userId int) ([]Post, error) {
	var posts []Post
	rows, err := db.Query(`WITH RECURSIVE submitted_comments AS (
								SELECT 1 as lev, id, text, votes, created_at, parent_id, author_id, RIGHT('0000' || id::VARCHAR, 4) || ' ' AS skey
								FROM posts WHERE author_id=$1 AND parent_id IS NOT NULL
								UNION ALL
								SELECT lev + 1, posts.id, posts.text, posts.votes, posts.created_at, posts.parent_id, posts.author_id,
								skey || RIGHT('0000' || (9999-posts.votes)::VARCHAR, 4) ||
								RIGHT('0000000000' || (EXTRACT(EPOCH FROM (NOW() - posts.created_at))::INTEGER)::VARCHAR, 10) || ' '
								FROM posts JOIN submitted_comments ON posts.parent_id = submitted_comments.id
							),
							comments AS (
								SELECT * FROM submitted_comments LIMIT $3
							),
							user_votes AS (
								SELECT user_id, post_id FROM votes WHERE user_id=$2
							),
							user_flags AS (
								SELECT user_id, post_id FROM flags WHERE user_id=$2
							)
							SELECT comments.lev, comments.id, comments.text, comments.votes, comments.created_at, 
								COALESCE(comments.parent_id, 0), username, COALESCE(user_votes.user_id, 0), 
								COALESCE(user_flags.user_id, 0), author_id
							FROM comments
							LEFT JOIN users ON comments.author_id=users.id
							LEFT JOIN user_votes ON comments.id=user_votes.post_id
							LEFT JOIN user_flags ON comments.id=user_flags.post_id
							ORDER BY skey
							`, authorId, userId, LIMIT*4)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.Lev, &post.ID, &post.Text, &post.Votes, &post.CreatedAt,
			&post.ParentID, &post.Author, &post.VotedID, &post.FlaggedID, &post.AuthorID); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func GetSubmittedComments(authorId int) ([]Post, error) {
	var posts []Post
	rows, err := db.Query(`WITH RECURSIVE submitted_comments AS (
								SELECT 1 as lev, id, text, votes, created_at, parent_id, author_id, RIGHT('0000' || id::VARCHAR, 4) || ' ' AS skey
								FROM posts WHERE author_id=$1 AND parent_id IS NOT NULL
								UNION ALL
								SELECT lev + 1, posts.id, posts.text, posts.votes, posts.created_at, posts.parent_id, posts.author_id,
								skey || RIGHT('0000' || (9999-posts.votes)::VARCHAR, 4) ||
								RIGHT('0000000000' || (EXTRACT(EPOCH FROM (NOW() - posts.created_at))::INTEGER)::VARCHAR, 10) || ' '
								FROM posts JOIN submitted_comments ON posts.parent_id = submitted_comments.id
							),
							comments AS (
								SELECT * FROM submitted_comments LIMIT $2
							)
							SELECT comments.lev, comments.id, comments.text, comments.votes, comments.created_at, 
								COALESCE(comments.parent_id, 0), username, author_id
							FROM comments
							LEFT JOIN users ON comments.author_id=users.id
							ORDER BY skey
							`, authorId, LIMIT*4)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.Lev, &post.ID, &post.Text, &post.Votes, &post.CreatedAt,
			&post.ParentID, &post.Author, &post.AuthorID); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (p *Post) GetAuth(currentUserID int) error {
	query := `SELECT title, url, text, votes, author_id, COALESCE(parent_id, 0), created_at, username, 
				COALESCE(votes.user_id, 0), COALESCE(flags.user_id, 0)
				FROM posts
				LEFT JOIN users ON posts.author_id=users.id
				LEFT JOIN votes ON posts.id=votes.post_id AND votes.user_id=$2
				LEFT JOIN flags ON posts.id=flags.post_id AND flags.user_id=$2
				WHERE posts.id=$1
			`
	err := db.QueryRow(query, p.ID, currentUserID).Scan(&p.Title, &p.Url, &p.Text, &p.Votes, &p.AuthorID, &p.ParentID,
		&p.CreatedAt, &p.Author, &p.VotedID, &p.FlaggedID)
	return err
}

func (p *Post) Get() error {
	query := `SELECT title, url, text, votes, author_id, COALESCE(parent_id, 0), created_at, username
				FROM posts
				LEFT JOIN users ON posts.author_id=users.id
				WHERE posts.id=$1
			`
	err := db.QueryRow(query, p.ID).Scan(&p.Title, &p.Url, &p.Text, &p.Votes, &p.AuthorID, &p.ParentID,
		&p.CreatedAt, &p.Author)
	return err
}

func (p *Post) GetAuthComments(userId int) ([]Post, error) {
	var posts []Post
	rows, err := db.Query(`WITH RECURSIVE submitted_comments AS (
								SELECT 1 as lev, id, text, votes, created_at, parent_id, author_id, RIGHT('0000' || id::VARCHAR, 4) || ' ' AS skey
								FROM posts WHERE parent_id=$1
								UNION ALL
								SELECT lev + 1, posts.id, posts.text, posts.votes, posts.created_at, posts.parent_id, posts.author_id,
								skey || RIGHT('0000' || (9999-posts.votes)::VARCHAR, 4) ||
								RIGHT('0000000000' || (EXTRACT(EPOCH FROM (NOW() - posts.created_at))::INTEGER)::VARCHAR, 10) || ' '
								FROM posts JOIN submitted_comments ON posts.parent_id = submitted_comments.id
							),
							comments AS (
								SELECT * FROM submitted_comments LIMIT $3
							),
							user_votes AS (
								SELECT user_id, post_id FROM votes WHERE user_id=$2
							),
							user_flags AS (
								SELECT user_id, post_id FROM flags WHERE user_id=$2
							)
							SELECT comments.lev, comments.id, comments.text, comments.votes, comments.created_at, 
								COALESCE(comments.parent_id, 0), username, COALESCE(user_votes.user_id, 0), 
								COALESCE(user_flags.user_id, 0), author_id
							FROM comments
							LEFT JOIN users ON comments.author_id=users.id
							LEFT JOIN user_votes ON comments.id=user_votes.post_id
							LEFT JOIN user_flags ON comments.id=user_flags.post_id
							ORDER BY skey
							`, p.ID, userId, LIMIT*4)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.Lev, &post.ID, &post.Text, &post.Votes, &post.CreatedAt,
			&post.ParentID, &post.Author, &post.VotedID, &post.FlaggedID, &post.AuthorID); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (p *Post) GetComments() ([]Post, error) {
	var posts []Post
	rows, err := db.Query(`WITH RECURSIVE submitted_comments AS (
								SELECT 1 as lev, id, text, votes, created_at, parent_id, author_id, RIGHT('0000' || id::VARCHAR, 4) || ' ' AS skey
								FROM posts WHERE parent_id=$1
								UNION ALL
								SELECT lev + 1, posts.id, posts.text, posts.votes, posts.created_at, posts.parent_id, posts.author_id,
								skey || RIGHT('0000' || (9999-posts.votes)::VARCHAR, 4) ||
								RIGHT('0000000000' || (EXTRACT(EPOCH FROM (NOW() - posts.created_at))::INTEGER)::VARCHAR, 10) || ' '
								FROM posts JOIN submitted_comments ON posts.parent_id = submitted_comments.id
							),
							comments AS (
								SELECT * FROM submitted_comments LIMIT $2
							)
							SELECT comments.lev, comments.id, comments.text, comments.votes, comments.created_at, 
								COALESCE(comments.parent_id, 0), username, author_id
							FROM comments
							LEFT JOIN users ON comments.author_id=users.id
							ORDER BY skey
							`, p.ID, LIMIT*4)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.Lev, &post.ID, &post.Text, &post.Votes, &post.CreatedAt,
			&post.ParentID, &post.Author, &post.AuthorID); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}
