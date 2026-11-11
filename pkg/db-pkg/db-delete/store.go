package dbdelete

import (
	"context"

	dbresult "life_easy_api/pkg/db-pkg/db-result"

	"github.com/jmoiron/sqlx"
)

func Entry(
	ctx context.Context,
	q sqlx.ExtContext,
	userID, id string,
) error {
	result, err := q.ExecContext(ctx, `
		DELETE FROM entries
		WHERE id = $1
		  AND owner_id = $2
		  AND source_bill_id IS NULL
		  AND NOT (kind = 'bill' AND done = TRUE)
	`, id, userID)

	return dbresult.RequireAffected(result, err)
}

func Member(
	ctx context.Context,
	q sqlx.ExtContext,
	ownerID, familyID, userID string,
) error {
	result, err := q.ExecContext(ctx, `
		DELETE FROM family_members m
		USING families f
		WHERE m.family_id = f.id
		  AND f.id = $1
		  AND f.owner_id = $2
		  AND m.user_id = $3
		  AND m.role = 'member'
	`, familyID, ownerID, userID)

	return dbresult.RequireAffected(result, err)
}
