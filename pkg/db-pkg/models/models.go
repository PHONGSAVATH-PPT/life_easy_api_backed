package models

import (
	"errors"
	"time"
)

var (
	ErrInvalid      = errors.New("invalid request")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

type User struct {
	ID           string `db:"id" json:"id"`
	Email        string `db:"email" json:"email"`
	Name         string `db:"name" json:"name"`
	PasswordHash string `db:"password_hash" json:"-"`
}

type Entry struct {
	ID           string    `db:"id" json:"id"`
	OwnerID      string    `db:"owner_id" json:"owner_id"`
	FamilyID     *string   `db:"family_id" json:"family_id"`
	Kind         string    `db:"kind" json:"kind"`
	Title        string    `db:"title" json:"title"`
	DueAt        time.Time `db:"due_at" json:"due_at"`
	AmountMinor  int64     `db:"amount_minor" json:"amount_minor"`
	Currency     string    `db:"currency" json:"currency"`
	Done         bool      `db:"done" json:"done"`
	Shared       bool      `db:"shared" json:"shared"`
	Note         string    `db:"note" json:"note"`
	AssigneeID   *string   `db:"assignee_id" json:"assignee_id"`
	Recurrence   string    `db:"recurrence" json:"recurrence"`
	Reminder     bool      `db:"reminder_enabled" json:"reminder_enabled"`
	SourceBillID *string   `db:"source_bill_id" json:"source_bill_id"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type Member struct {
	ID   string `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
	Role string `db:"role" json:"role"`
}

type Workspace struct {
	FamilyID *string  `json:"family_id"`
	Entries  []Entry  `json:"entries"`
	Members  []Member `json:"members"`
}

type Session struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	User        User   `json:"user"`
}
