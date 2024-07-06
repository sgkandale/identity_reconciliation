package database_test

import (
	"context"
	"testing"
	"time"

	"identity/config"
	"identity/database"
	"identity/database/postgres"
)

var dbConn database.Database

func init() {
	dbConn = postgres.New(
		context.Background(),
		&config.DatabaseConfig{
			Type:    "postgres",
			Uri:     "",
			Timeout: 30,
		},
	)
}

func TestNewPostgres(t *testing.T) {
	dbConn := postgres.New(
		context.Background(),
		&config.DatabaseConfig{
			Type:    "postgres",
			Uri:     "",
			Timeout: 30,
		},
	)
	if dbConn == nil {
		t.Error("Failed to create postgres connection")
	}
}

func TestPutContact(t *testing.T) {
	dbConn.PutContact(
		context.Background(),
		&database.Contact{
			PhoneNumber:    "1234",
			Email:          "abcd",
			LinkedId:       3,
			LinkPrecedence: database.ContactLinkPrecedence_Primary,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	)
}

func TestFindRecentContact(t *testing.T) {
	contact, err := dbConn.FindRecentContact(
		context.Background(),
		&database.Contact{
			PhoneNumber: "1234",
			Email:       "abcd",
		},
	)
	if err != nil {
		t.Error(err)
	}
	log.Printf("contact : %+v", contact)
}

func TestFindContactByEmail(t *testing.T) {
	contact, err := dbConn.FindContactByEmail(
		context.Background(),
		"abcd",
	)
	if err != nil {
		t.Error(err)
	}
	log.Printf("contact : %+v", contact)
}

func TestFindContactByPhone(t *testing.T) {
	contact, err := dbConn.FindContactByPhone(
		context.Background(),
		"1234",
	)
	if err != nil {
		t.Error(err)
	}
	log.Printf("contact : %+v", contact)
}
