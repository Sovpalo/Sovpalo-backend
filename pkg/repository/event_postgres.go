package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/jackc/pgx/v5"
)

func (r *EventPostgres) CreateEvent(event model.Event) (int64, error) {
	ctx := context.Background()

	if event.CompanyID != nil {
		var isMember bool
		err := r.pool.QueryRow(ctx,
			"SELECT EXISTS (SELECT 1 FROM company_members WHERE company_id = $1 AND user_id = $2)",
			*event.CompanyID, event.CreatedBy,
		).Scan(&isMember)
		if err != nil {
			return 0, err
		}
		if !isMember {
			return 0, errors.New("user is not a member of the company")
		}
	}

	query := `
		INSERT INTO events (group_id, company_id, created_by, title, description, start_time, end_time, place_name, place_link, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'pending')
		RETURNING id
	`
	var id int64
	err := r.pool.QueryRow(
		ctx,
		query,
		event.GroupID,
		event.CompanyID,
		event.CreatedBy,
		event.Title,
		event.Description,
		event.StartTime,
		event.EndTime,
		event.PlaceName,
		event.PlaceLink,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *EventPostgres) GetEvent(eventID int64, userID int64) (model.Event, error) {
	ctx := context.Background()
	var event model.Event
	query := `
		SELECT e.id, e.group_id, e.company_id, e.created_by, e.title, e.description, e.start_time, e.end_time,
		       e.place_name, e.place_link, e.status, e.created_at, e.updated_at
		FROM events e
		LEFT JOIN company_members cm ON cm.company_id = e.company_id AND cm.user_id = $2
		WHERE e.id = $1
		  AND (
		    (e.company_id IS NOT NULL AND cm.user_id IS NOT NULL)
		    OR (e.company_id IS NULL AND e.created_by = $2)
		  )
	`
	err := r.pool.QueryRow(ctx, query, eventID, userID).Scan(
		&event.ID,
		&event.GroupID,
		&event.CompanyID,
		&event.CreatedBy,
		&event.Title,
		&event.Description,
		&event.StartTime,
		&event.EndTime,
		&event.PlaceName,
		&event.PlaceLink,
		&event.Status,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		return model.Event{}, err
	}
	return event, nil
}

func (r *EventPostgres) ListEvents(userID int64) ([]model.Event, error) {
	ctx := context.Background()
	query := `
		SELECT DISTINCT e.id, e.group_id, e.company_id, e.created_by, e.title, e.description, e.start_time, e.end_time,
		       e.place_name, e.place_link, e.status, e.created_at, e.updated_at
		FROM events e
		LEFT JOIN company_members cm ON cm.company_id = e.company_id AND cm.user_id = $1
		WHERE (e.company_id IS NOT NULL AND cm.user_id IS NOT NULL)
		   OR (e.company_id IS NULL AND e.created_by = $1)
		ORDER BY e.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var event model.Event
		if err := rows.Scan(
			&event.ID,
			&event.GroupID,
			&event.CompanyID,
			&event.CreatedBy,
			&event.Title,
			&event.Description,
			&event.StartTime,
			&event.EndTime,
			&event.PlaceName,
			&event.PlaceLink,
			&event.Status,
			&event.CreatedAt,
			&event.UpdatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (r *EventPostgres) ListCompanyEvents(companyID int64, userID int64) ([]model.Event, error) {
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
		SELECT e.id, e.group_id, e.company_id, e.created_by, e.title, e.description, e.start_time, e.end_time,
		       e.place_name, e.place_link, e.status, e.created_at, e.updated_at
		FROM events e
		WHERE e.company_id = $1
		ORDER BY e.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var event model.Event
		if err := rows.Scan(
			&event.ID,
			&event.GroupID,
			&event.CompanyID,
			&event.CreatedBy,
			&event.Title,
			&event.Description,
			&event.StartTime,
			&event.EndTime,
			&event.PlaceName,
			&event.PlaceLink,
			&event.Status,
			&event.CreatedAt,
			&event.UpdatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (r *EventPostgres) UpdateEvent(eventID int64, userID int64, input model.EventUpdateInput) error {
	ctx := context.Background()

	setParts := make([]string, 0, 5)
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
	if input.StartTime != nil {
		setParts = append(setParts, fmt.Sprintf("start_time = $%d", argID))
		args = append(args, *input.StartTime)
		argID++
	}
	if input.EndTime != nil {
		setParts = append(setParts, fmt.Sprintf("end_time = $%d", argID))
		args = append(args, *input.EndTime)
		argID++
	}
	if input.CompanyID != nil {
		var isMember bool
		err := r.pool.QueryRow(ctx,
			"SELECT EXISTS (SELECT 1 FROM company_members WHERE company_id = $1 AND user_id = $2)",
			*input.CompanyID, userID,
		).Scan(&isMember)
		if err != nil {
			return err
		}
		if !isMember {
			return errors.New("user is not a member of the company")
		}

		setParts = append(setParts, fmt.Sprintf("company_id = $%d", argID))
		args = append(args, *input.CompanyID)
		argID++
	}

	if len(setParts) == 0 {
		return errors.New("no fields to update")
	}

	setParts = append(setParts, "updated_at = NOW()")
	query := fmt.Sprintf(
		"UPDATE events SET %s WHERE id = $%d AND created_by = $%d",
		strings.Join(setParts, ", "),
		argID,
		argID+1,
	)
	args = append(args, eventID, userID)

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *EventPostgres) DeleteEvent(eventID int64, userID int64) error {
	ctx := context.Background()
	query := "DELETE FROM events WHERE id = $1 AND created_by = $2"
	tag, err := r.pool.Exec(ctx, query, eventID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
