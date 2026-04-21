package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/jackc/pgx/v5"
)

func (r *IdeaPostgres) CreateCompanyIdea(companyID int64, userID int64, input model.IdeaCreateInput) (int64, error) {
	ctx := context.Background()

	var isMember bool
	if err := r.pool.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM company_members WHERE company_id = $1 AND user_id = $2)",
		companyID, userID,
	).Scan(&isMember); err != nil {
		return 0, err
	}
	if !isMember {
		return 0, errors.New("user is not a member of the company")
	}

	query := `
		INSERT INTO ideas (company_id, created_by, title, description, photo_url, source)
		VALUES ($1, $2, $3, $4, $5, 'manual')
		RETURNING id
	`
	var id int64
	if err := r.pool.QueryRow(ctx, query, companyID, userID, input.Title, input.Description, input.PhotoURL).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *IdeaPostgres) ListCompanyIdeas(companyID int64, userID int64) ([]model.IdeaView, error) {
	ctx := context.Background()

	var isMember bool
	if err := r.pool.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM company_members WHERE company_id = $1 AND user_id = $2)",
		companyID, userID,
	).Scan(&isMember); err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("user is not a member of the company")
	}

	query := `
		SELECT i.id,
		       i.title,
		       i.description,
		       i.photo_url,
		       i.company_id,
		       i.created_by,
		       u.username,
		       COALESCE(lc.likes_count, 0) AS likes_count,
		       EXISTS (
		           SELECT 1 FROM idea_likes il
		           WHERE il.idea_id = i.id AND il.user_id = $2
		       ) AS liked_by_current
		FROM ideas i
		JOIN users u ON u.id = i.created_by
		LEFT JOIN (
		    SELECT idea_id, COUNT(*) AS likes_count
		    FROM idea_likes
		    GROUP BY idea_id
		) lc ON lc.idea_id = i.id
		WHERE i.company_id = $1
		ORDER BY i.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, companyID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ideas []model.IdeaView
	for rows.Next() {
		var idea model.IdeaView
		if err := rows.Scan(
			&idea.ID,
			&idea.Title,
			&idea.Description,
			&idea.PhotoURL,
			&idea.CompanyID,
			&idea.CreatedBy,
			&idea.CreatedByUsername,
			&idea.LikesCount,
			&idea.LikedByCurrent,
		); err != nil {
			return nil, err
		}
		ideas = append(ideas, idea)
	}
	return ideas, rows.Err()
}

func (r *IdeaPostgres) GetCompanyIdea(companyID int64, userID int64, ideaID int64) (model.IdeaView, error) {
	ctx := context.Background()

	var isMember bool
	if err := r.pool.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM company_members WHERE company_id = $1 AND user_id = $2)",
		companyID, userID,
	).Scan(&isMember); err != nil {
		return model.IdeaView{}, err
	}
	if !isMember {
		return model.IdeaView{}, errors.New("user is not a member of the company")
	}

	query := `
		SELECT i.id,
		       i.title,
		       i.description,
		       i.photo_url,
		       i.company_id,
		       i.created_by,
		       u.username,
		       COALESCE(lc.likes_count, 0) AS likes_count,
		       EXISTS (
		           SELECT 1 FROM idea_likes il
		           WHERE il.idea_id = i.id AND il.user_id = $3
		       ) AS liked_by_current
		FROM ideas i
		JOIN users u ON u.id = i.created_by
		LEFT JOIN (
		    SELECT idea_id, COUNT(*) AS likes_count
		    FROM idea_likes
		    GROUP BY idea_id
		) lc ON lc.idea_id = i.id
		WHERE i.company_id = $1 AND i.id = $2
	`
	var idea model.IdeaView
	err := r.pool.QueryRow(ctx, query, companyID, ideaID, userID).Scan(
		&idea.ID,
		&idea.Title,
		&idea.Description,
		&idea.PhotoURL,
		&idea.CompanyID,
		&idea.CreatedBy,
		&idea.CreatedByUsername,
		&idea.LikesCount,
		&idea.LikedByCurrent,
	)
	if err != nil {
		return model.IdeaView{}, err
	}
	return idea, nil
}

func (r *IdeaPostgres) UpdateCompanyIdea(companyID int64, userID int64, ideaID int64, input model.IdeaUpdateInput) error {
	ctx := context.Background()

	var isMember bool
	if err := r.pool.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM company_members WHERE company_id = $1 AND user_id = $2)",
		companyID, userID,
	).Scan(&isMember); err != nil {
		return err
	}
	if !isMember {
		return errors.New("user is not a member of the company")
	}

	setParts := make([]string, 0, 4)
	args := make([]interface{}, 0, 6)
	argID := 1

	if input.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argID))
		args = append(args, *input.Title)
		argID++
	}
	if input.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argID))
		args = append(args, *input.Description)
		argID++
	}
	if input.PhotoURL != nil {
		setParts = append(setParts, fmt.Sprintf("photo_url = $%d", argID))
		args = append(args, *input.PhotoURL)
		argID++
	}

	if len(setParts) == 0 {
		return errors.New("no fields to update")
	}

	setParts = append(setParts, "updated_at = NOW()")
	query := fmt.Sprintf(
		"UPDATE ideas SET %s WHERE id = $%d AND company_id = $%d AND created_by = $%d",
		strings.Join(setParts, ", "),
		argID,
		argID+1,
		argID+2,
	)
	args = append(args, ideaID, companyID, userID)

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *IdeaPostgres) LikeCompanyIdea(companyID int64, userID int64, ideaID int64) error {
	ctx := context.Background()

	var isMember bool
	if err := r.pool.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM company_members WHERE company_id = $1 AND user_id = $2)",
		companyID, userID,
	).Scan(&isMember); err != nil {
		return err
	}
	if !isMember {
		return errors.New("user is not a member of the company")
	}

	var ideaCompanyID int64
	if err := r.pool.QueryRow(ctx, "SELECT company_id FROM ideas WHERE id = $1", ideaID).Scan(&ideaCompanyID); err != nil {
		return err
	}
	if ideaCompanyID != companyID {
		return pgx.ErrNoRows
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO idea_likes (idea_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (idea_id, user_id) DO NOTHING
	`, ideaID, userID)
	return err
}

func (r *IdeaPostgres) UnlikeCompanyIdea(companyID int64, userID int64, ideaID int64) error {
	ctx := context.Background()

	var isMember bool
	if err := r.pool.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM company_members WHERE company_id = $1 AND user_id = $2)",
		companyID, userID,
	).Scan(&isMember); err != nil {
		return err
	}
	if !isMember {
		return errors.New("user is not a member of the company")
	}

	var ideaCompanyID int64
	if err := r.pool.QueryRow(ctx, "SELECT company_id FROM ideas WHERE id = $1", ideaID).Scan(&ideaCompanyID); err != nil {
		return err
	}
	if ideaCompanyID != companyID {
		return pgx.ErrNoRows
	}

	_, err := r.pool.Exec(ctx, "DELETE FROM idea_likes WHERE idea_id = $1 AND user_id = $2", ideaID, userID)
	return err
}
