package dbupdate

import (
	"context"

	dbresult "life_easy_api/pkg/db-pkg/db-result"
	"life_easy_api/pkg/db-pkg/models"

	"github.com/jmoiron/sqlx"
)

func Entry(
	ctx context.Context,
	q sqlx.ExtContext,
	entry models.Entry,
) error {
	result, err := q.ExecContext(ctx, `
		UPDATE entries
		SET
			family_id = $3,
			kind = $4,
			title = $5,
			due_at = $6,
			amount_minor = $7,
			currency = $8,
			done = $9,
			shared = $10,
			note = $11,
			assignee_id = $12,
			recurrence = $13,
			reminder_enabled = $14,
			updated_at = NOW()
		WHERE id = $1
		  AND owner_id = $2
		  AND source_bill_id IS NULL
		  AND NOT (kind = 'bill' AND done = TRUE)
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
	)

	return dbresult.RequireAffected(result, err)
}

func ToggleTask(
	ctx context.Context,
	q sqlx.ExtContext,
	userID, id string,
) error {
	result, err := q.ExecContext(ctx, `
		UPDATE entries e
		SET done = NOT done, updated_at = NOW()
		WHERE e.id = $1
		  AND e.kind IN ('task', 'appointment')
		  AND (
				e.owner_id = $2
				OR (
					e.assignee_id = $2
					AND e.shared = TRUE
					AND EXISTS (
						SELECT 1
						FROM family_members m
						WHERE m.family_id = e.family_id
						  AND m.user_id = $2
					)
				)
		  )
	`, id, userID)

	return dbresult.RequireAffected(result, err)
}

func BillPaid(
	ctx context.Context,
	q sqlx.ExtContext,
	userID, id string,
) error {
	result, err := q.ExecContext(ctx, `
		UPDATE entries
		SET done = TRUE, updated_at = NOW()
		WHERE id = $1
		  AND owner_id = $2
		  AND kind = 'bill'
		  AND done = FALSE
	`, id, userID)

	return dbresult.RequireAffected(result, err)
}

func MemberEntriesPrivate(
	ctx context.Context,
	q sqlx.ExtContext,
	familyID, userID string,
) error {
	_, err := q.ExecContext(ctx, `
		UPDATE entries
		SET
			family_id = NULL,
			shared = FALSE,
			assignee_id = NULL,
			updated_at = NOW()
		WHERE family_id = $1
		  AND owner_id = $2
	`, familyID, userID)

	return err
}
