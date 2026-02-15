package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postStore struct {
	pool *pgxpool.Pool
}

func NewPostStore(pool *pgxpool.Pool) PostStore {
	return &postStore{pool: pool}
}

func (s *postStore) CreatePost(ctx context.Context, authorID, content string) (*models.Post, error) {
	var post models.Post
	err := s.pool.QueryRow(ctx,
		`INSERT INTO posts (author_id, content)
		 VALUES ($1, $2)
		 RETURNING id, author_id, content, created_at`,
		authorID, content,
	).Scan(&post.ID, &post.AuthorID, &post.Content, &post.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23514" {
			if pgErr.ConstraintName == "posts_content_max_length" {
				return nil, &models.APIError{Code: models.ErrCodeContentLong, Message: "Post content exceeds 140 characters"}
			}
			if pgErr.ConstraintName == "posts_content_not_empty" {
				return nil, &models.APIError{Code: models.ErrCodeValidation, Message: "Post content cannot be empty", Field: "content"}
			}
		}
		return nil, fmt.Errorf("create post: %w", err)
	}

	return &post, nil
}

func (s *postStore) GetPostByID(ctx context.Context, id string) (*models.Post, error) {
	var post models.Post
	var author models.User
	err := s.pool.QueryRow(ctx,
		`SELECT p.id, p.author_id, p.content, p.created_at,
		        u.id, u.username, u.display_name, u.bio, u.created_at
		 FROM posts p
		 JOIN users u ON p.author_id = u.id
		 WHERE p.id = $1`, id,
	).Scan(&post.ID, &post.AuthorID, &post.Content, &post.CreatedAt,
		&author.ID, &author.Username, &author.DisplayName, &author.Bio, &author.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &models.APIError{Code: models.ErrCodeNotFound, Message: "Post not found"}
		}
		return nil, fmt.Errorf("get post by id: %w", err)
	}

	post.Author = &author
	return &post, nil
}

func (s *postStore) GetTimeline(ctx context.Context, cursor time.Time, limit int) ([]models.Post, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT p.id, p.author_id, p.content, p.created_at,
		        u.id, u.username, u.display_name, u.bio, u.created_at
		 FROM posts p
		 JOIN users u ON p.author_id = u.id
		 WHERE p.created_at < $1
		 ORDER BY p.created_at DESC
		 LIMIT $2`, cursor, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get timeline: %w", err)
	}
	defer rows.Close()

	return scanPosts(rows)
}

func (s *postStore) GetUserPosts(ctx context.Context, userID string, cursor time.Time, limit int) ([]models.Post, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT p.id, p.author_id, p.content, p.created_at,
		        u.id, u.username, u.display_name, u.bio, u.created_at
		 FROM posts p
		 JOIN users u ON p.author_id = u.id
		 WHERE p.author_id = $1
		   AND p.created_at < $2
		 ORDER BY p.created_at DESC
		 LIMIT $3`, userID, cursor, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get user posts: %w", err)
	}
	defer rows.Close()

	return scanPosts(rows)
}

func scanPosts(rows pgx.Rows) ([]models.Post, error) {
	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var author models.User
		err := rows.Scan(
			&post.ID, &post.AuthorID, &post.Content, &post.CreatedAt,
			&author.ID, &author.Username, &author.DisplayName, &author.Bio, &author.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		post.Author = &author
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate posts: %w", err)
	}
	return posts, nil
}
