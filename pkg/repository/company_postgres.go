package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/jackc/pgx/v5"
)

func (r *CompanyPostgres) CreateCompany(company model.Company) (int64, error) {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var id int64
	query := "INSERT INTO companies (name, description, created_by) VALUES ($1, $2, $3) RETURNING id"
	if err := tx.QueryRow(ctx, query, company.Name, company.Description, company.CreatedBy).Scan(&id); err != nil {
		return 0, err
	}

	memberQuery := "INSERT INTO company_members (company_id, user_id, role) VALUES ($1, $2, 'owner')"
	if _, err := tx.Exec(ctx, memberQuery, id, company.CreatedBy); err != nil {
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *CompanyPostgres) GetCompany(companyID int64, userID int64) (model.Company, error) {
	ctx := context.Background()
	var company model.Company
	query := `
		SELECT c.id, c.name, c.description, c.created_by, c.created_at, c.updated_at
		FROM companies c
		JOIN company_members cm ON cm.company_id = c.id
		WHERE c.id = $1 AND cm.user_id = $2
	`
	err := r.pool.QueryRow(ctx, query, companyID, userID).Scan(
		&company.ID,
		&company.Name,
		&company.Description,
		&company.CreatedBy,
		&company.CreatedAt,
		&company.UpdatedAt,
	)
	if err != nil {
		return model.Company{}, err
	}
	return company, nil
}

func (r *CompanyPostgres) ListCompanies(userID int64) ([]model.Company, error) {
	ctx := context.Background()
	query := `
		SELECT c.id, c.name, c.description, c.created_by, c.created_at, c.updated_at
		FROM companies c
		JOIN company_members cm ON cm.company_id = c.id
		WHERE cm.user_id = $1
		ORDER BY c.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []model.Company
	for rows.Next() {
		var company model.Company
		if err := rows.Scan(
			&company.ID,
			&company.Name,
			&company.Description,
			&company.CreatedBy,
			&company.CreatedAt,
			&company.UpdatedAt,
		); err != nil {
			return nil, err
		}
		companies = append(companies, company)
	}
	return companies, rows.Err()
}

func (r *CompanyPostgres) UpdateCompany(companyID int64, userID int64, input model.CompanyUpdateInput) error {
	ctx := context.Background()
	setParts := make([]string, 0, 3)
	args := make([]interface{}, 0, 4)
	argID := 1

	if input.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argID))
		args = append(args, *input.Name)
		argID++
	}
	if input.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argID))
		args = append(args, *input.Description)
		argID++
	}

	if len(setParts) == 0 {
		return errors.New("no fields to update")
	}

	setParts = append(setParts, "updated_at = NOW()")
	query := fmt.Sprintf(
		"UPDATE companies SET %s WHERE id = $%d AND created_by = $%d",
		strings.Join(setParts, ", "),
		argID,
		argID+1,
	)
	args = append(args, companyID, userID)

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *CompanyPostgres) DeleteCompany(companyID int64, userID int64) error {
	ctx := context.Background()
	query := "DELETE FROM companies WHERE id = $1 AND created_by = $2"
	tag, err := r.pool.Exec(ctx, query, companyID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *CompanyPostgres) CreateInvitation(companyID int64, invitedBy int64, username string) (model.CompanyInvitation, error) {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.CompanyInvitation{}, err
	}
	defer tx.Rollback(ctx)

	var ownerID int64
	if err := tx.QueryRow(ctx, "SELECT created_by FROM companies WHERE id = $1", companyID).Scan(&ownerID); err != nil {
		return model.CompanyInvitation{}, err
	}
	if ownerID != invitedBy {
		return model.CompanyInvitation{}, errors.New("only company owner can invite")
	}

	var invitedUserID int64
	if err := tx.QueryRow(ctx, "SELECT id FROM users WHERE username = $1", username).Scan(&invitedUserID); err != nil {
		if err == sql.ErrNoRows || err == pgx.ErrNoRows {
			return model.CompanyInvitation{}, errors.New("user not found")
		}
		return model.CompanyInvitation{}, err
	}
	if invitedUserID == invitedBy {
		return model.CompanyInvitation{}, errors.New("cannot invite yourself")
	}

	var exists bool
	if err := tx.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM company_members WHERE company_id = $1 AND user_id = $2)",
		companyID, invitedUserID,
	).Scan(&exists); err != nil {
		return model.CompanyInvitation{}, err
	}
	if exists {
		return model.CompanyInvitation{}, errors.New("user already in company")
	}

	if err := tx.QueryRow(ctx,
		"SELECT EXISTS (SELECT 1 FROM company_invitations WHERE company_id = $1 AND invited_user_id = $2 AND status = 'pending')",
		companyID, invitedUserID,
	).Scan(&exists); err != nil {
		return model.CompanyInvitation{}, err
	}
	if exists {
		return model.CompanyInvitation{}, errors.New("invitation already sent")
	}

	var invitation model.CompanyInvitation
	inviteQuery := `
		INSERT INTO company_invitations (company_id, invited_user_id, invited_by, status)
		VALUES ($1, $2, $3, 'pending')
		RETURNING id, company_id, invited_user_id, invited_by, status, created_at
	`
	if err := tx.QueryRow(ctx, inviteQuery, companyID, invitedUserID, invitedBy).Scan(
		&invitation.ID,
		&invitation.CompanyID,
		&invitation.InvitedUserID,
		&invitation.InvitedBy,
		&invitation.Status,
		&invitation.CreatedAt,
	); err != nil {
		return model.CompanyInvitation{}, err
	}

	var companyName string
	if err := tx.QueryRow(ctx, "SELECT name FROM companies WHERE id = $1", companyID).Scan(&companyName); err != nil {
		return model.CompanyInvitation{}, err
	}
	var inviterUsername string
	if err := tx.QueryRow(ctx, "SELECT username FROM users WHERE id = $1", invitedBy).Scan(&inviterUsername); err != nil {
		return model.CompanyInvitation{}, err
	}

	notificationTitle := "Company invitation"
	notificationMessage := fmt.Sprintf("You were invited to %s by %s", companyName, inviterUsername)
	_, err = tx.Exec(ctx, `
		INSERT INTO notifications (user_id, type, title, message, related_entity_type, related_entity_id)
		VALUES ($1, 'company_invite', $2, $3, 'company_invitation', $4)
	`, invitedUserID, notificationTitle, notificationMessage, invitation.ID)
	if err != nil {
		return model.CompanyInvitation{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return model.CompanyInvitation{}, err
	}
	return invitation, nil
}

func (r *CompanyPostgres) ListInvitations(userID int64) ([]model.CompanyInvitationView, error) {
	ctx := context.Background()
	query := `
		SELECT ci.id,
		       ci.company_id,
		       c.name AS company_name,
		       ci.invited_by,
		       u.username AS invited_by_username,
		       ci.status,
		       ci.created_at
		FROM company_invitations ci
		JOIN companies c ON c.id = ci.company_id
		JOIN users u ON u.id = ci.invited_by
		WHERE ci.invited_user_id = $1 AND ci.status = 'pending'
		ORDER BY ci.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []model.CompanyInvitationView
	for rows.Next() {
		var invite model.CompanyInvitationView
		if err := rows.Scan(
			&invite.ID,
			&invite.CompanyID,
			&invite.CompanyName,
			&invite.InvitedBy,
			&invite.InvitedByUsername,
			&invite.Status,
			&invite.CreatedAt,
		); err != nil {
			return nil, err
		}
		invites = append(invites, invite)
	}
	return invites, rows.Err()
}

func (r *CompanyPostgres) AcceptInvitation(inviteID int64, userID int64) error {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var invitation model.CompanyInvitation
	query := `
		SELECT id, company_id, invited_user_id, invited_by, status, created_at, responded_at
		FROM company_invitations
		WHERE id = $1 AND invited_user_id = $2
	`
	if err := tx.QueryRow(ctx, query, inviteID, userID).Scan(
		&invitation.ID,
		&invitation.CompanyID,
		&invitation.InvitedUserID,
		&invitation.InvitedBy,
		&invitation.Status,
		&invitation.CreatedAt,
		&invitation.RespondedAt,
	); err != nil {
		return err
	}
	if invitation.Status != "pending" {
		return errors.New("invitation already handled")
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO company_members (company_id, user_id, role)
		VALUES ($1, $2, 'member')
		ON CONFLICT (company_id, user_id) DO NOTHING
	`, invitation.CompanyID, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE company_invitations
		SET status = 'accepted', responded_at = NOW()
		WHERE id = $1
	`, inviteID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *CompanyPostgres) DeclineInvitation(inviteID int64, userID int64) error {
	ctx := context.Background()
	query := `
		UPDATE company_invitations
		SET status = 'declined', responded_at = NOW()
		WHERE id = $1 AND invited_user_id = $2 AND status = 'pending'
	`
	tag, err := r.pool.Exec(ctx, query, inviteID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
