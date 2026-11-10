package schemas

import "time"

type Register struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=254"`
	Password string `json:"password" validate:"required,min=12,max=72"`
}

type Login struct {
	Email    string `json:"email" validate:"required,email,max=254"`
	Password string `json:"password" validate:"required,max=72"`
}

type UpsertEntry struct {
	ID           string    `json:"id"`
	FamilyID     *string   `json:"family_id" validate:"omitempty,uuid4"`
	Kind         string    `json:"kind" validate:"required,oneof=task appointment income expense bill document"`
	Title        string    `json:"title" validate:"required,max=200"`
	DueAt        time.Time `json:"due_at" validate:"required"`
	AmountMinor  int64     `json:"amount_minor" validate:"gte=0,lte=90000000000000"`
	Currency     string    `json:"currency" validate:"required,oneof=LAK THB"`
	Done         bool      `json:"done"`
	Shared       bool      `json:"shared"`
	Note         string    `json:"note" validate:"max=5000"`
	AssigneeID   *string   `json:"assignee_id" validate:"omitempty,uuid4"`
	Recurrence   string    `json:"recurrence" validate:"oneof=none monthly"`
	Reminder     bool      `json:"reminder_enabled"`
	SourceBillID *string   `json:"source_bill_id"`
}

type CreateFamily struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type AddMember struct {
	Email string `json:"email" validate:"required,email,max=254"`
}
