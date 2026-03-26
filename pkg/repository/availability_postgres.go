package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/jackc/pgx/v5"
)

func (r *AvailabilityPostgres) CreateAvailability(companyID int64, userID int64, input model.AvailabilityCreateInput) (int64, error) {
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
		INSERT INTO user_availability (user_id, company_id, start_time, end_time, note)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var id int64
	if err := r.pool.QueryRow(ctx, query, userID, companyID, input.StartTime, input.EndTime, input.Note).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *AvailabilityPostgres) ListAvailability(companyID int64, userID int64) ([]model.UserAvailability, error) {
	ctx := context.Background()
	query := `
		SELECT id, user_id, company_id, start_time, end_time, note, created_at, updated_at
		FROM user_availability
		WHERE company_id = $1 AND user_id = $2
		ORDER BY start_time
	`
	rows, err := r.pool.Query(ctx, query, companyID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.UserAvailability
	for rows.Next() {
		var item model.UserAvailability
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.CompanyID,
			&item.StartTime,
			&item.EndTime,
			&item.Note,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AvailabilityPostgres) ListCompanyAvailability(companyID int64, userID int64) ([]model.UserAvailability, error) {
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
		SELECT id, user_id, company_id, start_time, end_time, note, created_at, updated_at
		FROM user_availability
		WHERE company_id = $1
		ORDER BY user_id, start_time
	`
	rows, err := r.pool.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.UserAvailability
	for rows.Next() {
		var item model.UserAvailability
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.CompanyID,
			&item.StartTime,
			&item.EndTime,
			&item.Note,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AvailabilityPostgres) UpdateAvailability(companyID int64, userID int64, availabilityID int64, input model.AvailabilityCreateInput) error {
	ctx := context.Background()
	query := `
		UPDATE user_availability
		SET start_time = $1, end_time = $2, note = $3, updated_at = NOW()
		WHERE id = $4 AND company_id = $5 AND user_id = $6
	`
	tag, err := r.pool.Exec(ctx, query, input.StartTime, input.EndTime, input.Note, availabilityID, companyID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *AvailabilityPostgres) DeleteAvailability(companyID int64, userID int64, availabilityID int64) error {
	ctx := context.Background()
	query := `
		DELETE FROM user_availability
		WHERE id = $1 AND company_id = $2 AND user_id = $3
	`
	tag, err := r.pool.Exec(ctx, query, availabilityID, companyID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *AvailabilityPostgres) ListCompanyMemberIDs(companyID int64) ([]int64, error) {
	ctx := context.Background()
	query := "SELECT user_id FROM company_members WHERE company_id = $1"
	rows, err := r.pool.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *AvailabilityPostgres) ListAvailabilityInRange(companyID int64, start time.Time, end time.Time) ([]model.UserAvailability, error) {
	ctx := context.Background()
	query := `
		SELECT id, user_id, company_id, start_time, end_time, note, created_at, updated_at
		FROM user_availability
		WHERE company_id = $1
		  AND start_time < $3
		  AND end_time > $2
		ORDER BY user_id, start_time
	`
	rows, err := r.pool.Query(ctx, query, companyID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.UserAvailability
	for rows.Next() {
		var item model.UserAvailability
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.CompanyID,
			&item.StartTime,
			&item.EndTime,
			&item.Note,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
