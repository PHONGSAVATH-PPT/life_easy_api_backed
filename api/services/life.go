package services

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"time"

	"life_easy_api/api/schemas"
	dbdelete "life_easy_api/pkg/db-pkg/db-delete"
	dbinsert "life_easy_api/pkg/db-pkg/db-insert"
	dbselect "life_easy_api/pkg/db-pkg/db-select"
	dbupdate "life_easy_api/pkg/db-pkg/db-update"
	"life_easy_api/pkg/db-pkg/models"
	"life_easy_api/pkg/utils"
	check "life_easy_api/pkg/validator"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	DB        *sqlx.DB
	JWTSecret string
}

var entryIDPattern = regexp.MustCompile(
	`^[A-Za-z0-9_-]{1,100}$`,
)

func (s *Service) transaction(
	ctx context.Context,
	action func(*sqlx.Tx) error,
) error {
	tx, err := s.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := action(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) session(user models.User) (models.Session, error) {
	token, err := utils.GenerateToken(user.ID, s.JWTSecret)
	if err != nil {
		return models.Session{}, err
	}

	return models.Session{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(utils.Lifetime.Seconds()),
		User:        user,
	}, nil
}

func (s *Service) Register(
	ctx context.Context,
	input schemas.Register,
) (models.Session, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Name = strings.TrimSpace(input.Name)

	if err := check.Check(input); err != nil {
		return models.Session{}, err
	}

	// bcrypt จำกัด password ไม่เกิน 72 bytes
	if len(input.Password) > 72 {
		return models.Session{}, models.ErrInvalid
	}

	hash, err := utils.HashPassword(input.Password)
	if err != nil {
		return models.Session{}, err
	}

	user := models.User{
		ID:           uuid.NewString(),
		Email:        input.Email,
		Name:         input.Name,
		PasswordHash: hash,
	}

	if err := dbinsert.User(ctx, s.DB, user); err != nil {
		return models.Session{}, err
	}

	return s.session(user)
}

func (s *Service) Login(
	ctx context.Context,
	input schemas.Login,
) (models.Session, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	if err := check.Check(input); err != nil {
		return models.Session{}, models.ErrUnauthorized
	}

	if len(input.Password) > 72 {
		return models.Session{}, models.ErrUnauthorized
	}

	user, err := dbselect.UserByEmail(ctx, s.DB, input.Email)

	if errors.Is(err, sql.ErrNoRows) {
		return models.Session{}, models.ErrUnauthorized
	}
	if err != nil {
		return models.Session{}, err
	}

	if !utils.ComparePassword(user.PasswordHash, input.Password) {
		return models.Session{}, models.ErrUnauthorized
	}

	return s.session(user)
}

func (s *Service) Workspace(
	ctx context.Context,
	userID string,
) (models.Workspace, error) {
	// ให้รายการและสมาชิกมาจาก snapshot ฐานข้อมูลเดียวกัน
	tx, err := s.DB.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:  true,
	})
	if err != nil {
		return models.Workspace{}, err
	}
	defer tx.Rollback()

	result, err := dbselect.Workspace(ctx, tx, userID)
	if err != nil {
		return models.Workspace{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.Workspace{}, err
	}

	return result, nil
}

func (s *Service) SaveEntry(
	ctx context.Context,
	userID, id string,
	input schemas.EntryInput,
) error {
	input.Title = strings.TrimSpace(input.Title)

	if input.Recurrence == "" {
		input.Recurrence = "none"
	}

	if !entryIDPattern.MatchString(id) ||
		(input.ID != "" && input.ID != id) {
		return models.ErrInvalid
	}

	if err := check.Check(input); err != nil {
		return err
	}

	if input.Title == "" || input.SourceBillID != nil {
		return models.ErrInvalid
	}

	isTask := input.Kind == "task" || input.Kind == "appointment"
	isMoney := input.Kind == "income" ||
		input.Kind == "expense" ||
		input.Kind == "bill"

	if (isMoney && input.AmountMinor <= 0) ||
		(!isMoney && input.AmountMinor != 0) {
		return models.ErrInvalid
	}

	if input.Done && !isTask {
		return models.ErrInvalid
	}

	if input.AssigneeID != nil {
		if !isTask {
			return models.ErrInvalid
		}
		if !input.Shared && *input.AssigneeID != userID {
			return models.ErrForbidden
		}
	}

	if !isTask && input.Kind != "bill" &&
		input.Recurrence != "none" {
		return models.ErrInvalid
	}

	entry := models.Entry{
		ID:          id,
		OwnerID:     userID,
		FamilyID:    input.FamilyID,
		Kind:        input.Kind,
		Title:       input.Title,
		DueAt:       input.DueAt,
		AmountMinor: input.AmountMinor,
		Currency:    input.Currency,
		Done:        input.Done,
		Shared:      input.Shared,
		Note:        input.Note,
		AssigneeID:  input.AssigneeID,
		Recurrence:  input.Recurrence,
		Reminder:    input.Reminder,
	}

	return s.transaction(ctx, func(tx *sqlx.Tx) error {
		if entry.Shared {
			familyID, err := dbselect.FamilyForShare(
				ctx, tx, userID,
			)
			if errors.Is(err, sql.ErrNoRows) {
				return models.ErrForbidden
			}
			if err != nil {
				return err
			}

			if entry.FamilyID != nil &&
				*entry.FamilyID != familyID {
				return models.ErrForbidden
			}

			entry.FamilyID = &familyID
		} else {
			entry.FamilyID = nil
		}

		current, err := dbselect.EntryForUpdate(ctx, tx, id)

		if errors.Is(err, sql.ErrNoRows) {
			return dbinsert.Entry(ctx, tx, entry)
		}
		if err != nil {
			return err
		}

		if current.OwnerID != userID {
			return sql.ErrNoRows
		}

		if current.SourceBillID != nil ||
			(current.Kind == "bill" && current.Done) {
			return sql.ErrNoRows
		}

		return dbupdate.Entry(ctx, tx, entry)
	})
}

func (s *Service) RemoveEntry(
	ctx context.Context,
	userID, id string,
) error {
	if !entryIDPattern.MatchString(id) {
		return models.ErrInvalid
	}
	return dbdelete.Entry(ctx, s.DB, userID, id)
}

func (s *Service) ToggleTask(
	ctx context.Context,
	userID, id string,
) error {
	if !entryIDPattern.MatchString(id) {
		return models.ErrInvalid
	}
	return dbupdate.ToggleTask(ctx, s.DB, userID, id)
}

func (s *Service) PayBill(
	ctx context.Context,
	userID, id string,
) error {
	if !entryIDPattern.MatchString(id) {
		return models.ErrInvalid
	}

	return s.transaction(ctx, func(tx *sqlx.Tx) error {
		bill, err := dbselect.EntryForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}

		if bill.OwnerID != userID || bill.Kind != "bill" {
			return sql.ErrNoRows
		}

		// ทำซ้ำบิลเดิมไม่สร้างรายจ่ายใหม่
		if bill.Done {
			return nil
		}

		sourceID := bill.ID

		expense := models.Entry{
			ID:           uuid.NewString(),
			OwnerID:      userID,
			FamilyID:     bill.FamilyID,
			Kind:         "expense",
			Title:        bill.Title,
			DueAt:        time.Now().UTC(),
			AmountMinor:  bill.AmountMinor,
			Currency:     bill.Currency,
			Shared:       bill.Shared,
			Note:         "Recorded from bill payment",
			Recurrence:   "none",
			SourceBillID: &sourceID,
		}

		if err := dbinsert.Entry(ctx, tx, expense); err != nil {
			return err
		}

		return dbupdate.BillPaid(ctx, tx, userID, id)
	})
}

func (s *Service) CreateFamily(
	ctx context.Context,
	userID string,
	input schemas.CreateFamily,
) (string, error) {
	input.Name = strings.TrimSpace(input.Name)

	if err := check.Check(input); err != nil {
		return "", err
	}

	id := uuid.NewString()

	err := s.transaction(ctx, func(tx *sqlx.Tx) error {
		if err := dbinsert.Family(
			ctx, tx, id, userID, input.Name,
		); err != nil {
			return err
		}
		return dbinsert.FamilyOwner(ctx, tx, id, userID)
	})

	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *Service) AddMember(
	ctx context.Context,
	ownerID, familyID string,
	input schemas.AddMember,
) error {
	if _, err := uuid.Parse(familyID); err != nil {
		return models.ErrInvalid
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	if err := check.Check(input); err != nil {
		return err
	}

	return dbinsert.Member(
		ctx, s.DB, ownerID, familyID, input.Email,
	)
}

func (s *Service) RemoveMember(
	ctx context.Context,
	ownerID, familyID, userID string,
) error {
	if _, err := uuid.Parse(familyID); err != nil {
		return models.ErrInvalid
	}
	if _, err := uuid.Parse(userID); err != nil {
		return models.ErrInvalid
	}

	return s.transaction(ctx, func(tx *sqlx.Tx) error {
		if err := dbdelete.Member(
			ctx, tx, ownerID, familyID, userID,
		); err != nil {
			return err
		}

		return dbupdate.MemberEntriesPrivate(
			ctx, tx, familyID, userID,
		)
	})
}
