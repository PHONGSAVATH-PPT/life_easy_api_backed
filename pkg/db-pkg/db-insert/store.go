package dbinsert

import (
	"context"

	dbresult "life_easy_api/pkg/db-pkg/db-result"
	"life_easy_api/pkg/db-pkg/models"

	"github.com/jmoiron/sqlx"
)

func User(
	ctx context.Context,
	q sqlx.ExtContext,
	user models.User,
) error {
	_, err := q.ExecContext(ctx, `
		INSERT INTO users (id, email, name, password_hash)
		VALUES ($1, $2, $3, $4)
	`, user.ID, user.Email, user.Name, user.PasswordHash)

	return err
}

func Entry(
	ctx context.Context,
	q sqlx.ExtContext,
	entry models.Entry,
) error {
	_, err := q.ExecContext(ctx, `
		INSERT INTO entries (
			id, owner_id, family_id, kind, title,
			due_at, amount_minor, currency, done, shared,
			note, assignee_id, recurrence,
			reminder_enabled, source_bill_id
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15
		)
	`,
		entry.ID,
		entry.OwnerID,
		entry.FamilyID,
		entry.Kind,
		entry.Title,
		entry.DueAt,
		entry.AmountMinor,
		entry.Currency,
		entry.Done,
		entry.Shared,
		entry.Note,
		entry.AssigneeID,
		entry.Recurrence,
		entry.Reminder,
		entry.SourceBillID,
	)

	return err
}

func Family(
	ctx context.Context,
	q sqlx.ExtContext,
	id, ownerID, name string,
) error {
	_, err := q.ExecContext(ctx, `
		INSERT INTO families (id, owner_id, name)
		VALUES ($1, $2, $3)
	`, id, ownerID, name)

	return err
}

func FamilyOwner(
	ctx context.Context,
	q sqlx.ExtContext,
	familyID, userID string,
) error {
	_, err := q.ExecContext(ctx, `
		INSERT INTO family_members (family_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, familyID, userID)

	return err
}

func Member(
	ctx context.Context,
	q sqlx.ExtContext,
	ownerID, familyID, email string,
) error {
	result, err := q.ExecContext(ctx, `
		INSERT INTO family_members (family_id, user_id, role)
		SELECT f.id, u.id, 'member'
		FROM families f
		JOIN users u ON u.email = $3
		WHERE f.id = $1
		  AND f.owner_id = $2
	`, familyID, ownerID, email)

	return dbresult.RequireAffected(result, err)
}
