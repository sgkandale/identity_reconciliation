package database_test

import (
	"context"
	"log"
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
	err := dbConn.PutContact(
		context.Background(),
		&database.Contact{
			PhoneNumber:    "123456",
			Email:          "mcfly@hillvalley.edu",
			LinkedId:       nil,
			LinkPrecedence: database.ContactLinkPrecedence_Secondary,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	)
	if err != nil {
		t.Error(err)
	}
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
		return
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
		return
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
		return
	}
	log.Printf("contact : %+v", contact)
}

func TestFindAllContacts(t *testing.T) {
	contacts, err := dbConn.FindAllContacts(
		context.Background(),
		"abcd",
		"1234",
	)
	if err != nil {
		t.Error(err)
		return
	}

	for _, eachContact := range contacts {
		log.Printf("contact : %+v", eachContact)
	}
	log.Printf("total contacts : %d", len(contacts))
}
