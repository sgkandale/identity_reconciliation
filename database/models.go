package database

import "time"

type ContactLinkPrecedence string

const (
	ContactLinkPrecedence_Primary   = "primary"
	ContactLinkPrecedence_Secondary = "secondary"
)

type Contact struct {
	Id             int64                 `json:"id,omitempty"`
	PhoneNumber    string                `json:"phoneNumber,omitempty"`
	Email          string                `json:"email,omitempty"`
	LinkedId       int64                 `json:"linkedId,omitempty"`
	LinkPrecedence ContactLinkPrecedence `json:"linkPrecedence,omitempty"`
	CreatedAt      time.Time             `json:"createdAt,omitempty"`
	UpdatedAt      time.Time             `json:"updatedAt,omitempty"`
	DeletedAt      *time.Time            `json:"deletedAt,omitempty"`
}
