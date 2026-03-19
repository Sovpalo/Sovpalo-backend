package repository

import (
	"context"
	"errors"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
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
		INSERT INTO ideas (group_id, company_id, created_by, title, description, source)
		VALUES (NULL, $1, $2, $3, $4, 'manual')
		RETURNING id
	`
	var id int64
	if err := r.pool.QueryRow(ctx, query, companyID, userID, input.Title, input.Description).Scan(&id); err != nil {
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
		SELECT i.id, i.title, i.description, i.company_id, i.created_by, u.username
		FROM ideas i
		JOIN users u ON u.id = i.created_by
		WHERE i.company_id = $1
		ORDER BY i.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, companyID)
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
			&idea.CompanyID,
			&idea.CreatedBy,
			&idea.CreatedByUsername,
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
		SELECT i.id, i.title, i.description, i.company_id, i.created_by, u.username
		FROM ideas i
		JOIN users u ON u.id = i.created_by
		WHERE i.company_id = $1 AND i.id = $2
	`
	var idea model.IdeaView
	err := r.pool.QueryRow(ctx, query, companyID, ideaID).Scan(
		&idea.ID,
		&idea.Title,
		&idea.Description,
		&idea.CompanyID,
		&idea.CreatedBy,
		&idea.CreatedByUsername,
	)
	if err != nil {
		return model.IdeaView{}, err
	}
	return idea, nil
}
