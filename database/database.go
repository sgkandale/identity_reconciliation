package database

import "context"

type Database interface {
	PutContact(ctx context.Context, contact *Contact) error
	FindRecentContact(ctx context.Context, childContact *Contact) (*Contact, error)
	FindContactByEmail(ctx context.Context, email string) (*Contact, error)
	FindContactByPhone(ctx context.Context, phone string) (*Contact, error)
}
