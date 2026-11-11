package dbselect

import (
	"context"
	"database/sql"
	"errors"

	"life_easy_api/pkg/db-pkg/models"

	"github.com/jmoiron/sqlx"
)

func UserByEmail(
	ctx context.Context,
	q sqlx.ExtContext,
	email string,
) (models.User, error) {
	var user models.User

	err := sqlx.GetContext(ctx, q, &user, `
		SELECT id, email, name, password_hash
		FROM users
		WHERE email = $1
	`, email)

	return user, err
}

func EntryForUpdate(
	ctx context.Context,
	tx *sqlx.Tx,
	id string,
) (models.Entry, error) {
	var entry models.Entry

	err := tx.GetContext(ctx, &entry, `
		SELECT *
		FROM entries
		WHERE id = $1
		FOR UPDATE
	`, id)

	return entry, err
}

func FamilyForShare(
	ctx context.Context,
	tx *sqlx.Tx,
	userID string,
) (string, error) {
	var familyID string

	err := tx.GetContext(ctx, &familyID, `
		SELECT family_id
		FROM family_members
		WHERE user_id = $1
		FOR SHARE
	`, userID)

	return familyID, err
}

func Workspace(
	ctx context.Context,
	q sqlx.ExtContext,
	userID string,
) (models.Workspace, error) {
	result := models.Workspace{
		Entries: []models.Entry{},
		Members: []models.Member{},
	}

	var familyID string

	err := sqlx.GetContext(ctx, q, &familyID, `
		SELECT family_id
		FROM family_members
		WHERE user_id = $1
	`, userID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return result, err
	}

	if err == nil {
		result.FamilyID = &familyID
	}

	err = sqlx.SelectContext(ctx, q, &result.Entries, `
		SELECT e.*
		FROM entries e
		WHERE e.owner_id = $1
		   OR (
				e.shared = TRUE
				AND EXISTS (
					SELECT 1
					FROM family_members m
					WHERE m.family_id = e.family_id
					  AND m.user_id = $1
				)
		   )
		ORDER BY e.due_at DESC, e.id
	`, userID)

	if err != nil {
		return result, err
	}

	err = sqlx.SelectContext(ctx, q, &result.Members, `
		SELECT
			u.id,
			u.name,
			COALESCE(m.role, 'owner') AS role
		FROM users u
		LEFT JOIN family_members m ON m.user_id = u.id
		WHERE u.id = $1
		   OR m.family_id = (
				SELECT family_id
				FROM family_members
				WHERE user_id = $1
		   )
		ORDER BY u.name
	`, userID)

	return result, err
}
