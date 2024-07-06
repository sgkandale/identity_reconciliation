package postgres

import (
	"context"
	"fmt"
	"sync"

	"identity/database"

	"github.com/jackc/pgx/v5"
)

func (c *Client) PutContact(ctx context.Context, contact *database.Contact) error {
	callCtx, cancelCallCtx := context.WithTimeout(ctx, c.timeout)
	defer cancelCallCtx()

	insertFields := ``
	insertPositions := ``
	insertPositionsCount := 0
	argumentValues := []interface{}{}

	// if phone number is present
	if contact.PhoneNumber != "" {
		argumentValues = append(argumentValues, contact.PhoneNumber)
		insertFields += `phoneNumber, `
		insertPositionsCount++
		insertPositions += fmt.Sprintf(`$%d, `, insertPositionsCount)
	}

	// if email is present
	if contact.Email != "" {
		argumentValues = append(argumentValues, contact.Email)
		insertFields += `email, `
		insertPositionsCount++
		insertPositions += fmt.Sprintf(`$%d, `, insertPositionsCount)
	}

	// if linked id is present
	if contact.LinkedId > 0 {
		argumentValues = append(argumentValues, contact.LinkedId)
		insertFields += `linkedId, `
		insertPositionsCount++
		insertPositions += fmt.Sprintf(`$%d, `, insertPositionsCount)
	}

	// append rest of the arguments
	argumentValues = append(argumentValues, contact.LinkPrecedence, contact.CreatedAt, contact.UpdatedAt)
	insertPositions += fmt.Sprintf(`$%d, $%d, $%d`, insertPositionsCount+1, insertPositionsCount+2, insertPositionsCount+3)

	_, err := c.Pool.Exec(
		callCtx,
		`INSERT INTO `+TableName_Contact+`
		(`+insertFields+` 
		linkPrecedence, createdAt, updatedAt)
		VALUES
		(`+insertPositions+`)`,
		argumentValues...,
	)
	if err != nil {
		return fmt.Errorf("inserting in postgres : %s", err.Error())
	}
	return nil
}

func (c Client) FindRecentContact(ctx context.Context, childContact *database.Contact) (*database.Contact, error) {
	// if email and phone are empty
	// return nil
	if childContact.Email == "" && childContact.PhoneNumber == "" {
		return nil, nil
	}

	wg := sync.WaitGroup{}
	var lastContactByPhone *database.Contact
	var lastContactByPhoneErr error
	var lastContactByEmail *database.Contact
	var lastContactByEmailErr error

	// if phone number is present
	// find last inserted contact by phone number
	if childContact.PhoneNumber != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lastContactByPhone, lastContactByPhoneErr = c.FindContactByPhone(ctx, childContact.PhoneNumber)
		}()
	}

	// if email is present
	// find last inserted contact by email
	if childContact.Email != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lastContactByEmail, lastContactByEmailErr = c.FindContactByEmail(ctx, childContact.Email)
		}()
	}

	wg.Wait()

	if lastContactByPhoneErr != nil && lastContactByPhoneErr != pgx.ErrNoRows {
		return nil, fmt.Errorf("finding contact by phone : %s", lastContactByPhoneErr.Error())
	}
	if lastContactByEmailErr != nil && lastContactByEmailErr != pgx.ErrNoRows {
		return nil, fmt.Errorf("finding contact by email : %s", lastContactByEmailErr.Error())
	}

	// if both are present, return the latest one
	if lastContactByPhone != nil && lastContactByEmail != nil {
		if lastContactByPhone.UpdatedAt.After(lastContactByEmail.UpdatedAt) {
			return lastContactByPhone, nil
		}
		return lastContactByEmail, nil
	}

	// if any one of the two is present, return that one
	if lastContactByPhone != nil {
		return lastContactByPhone, nil
	}
	if lastContactByEmail != nil {
		return lastContactByEmail, nil
	}

	return nil, nil
}

func (c Client) FindContactByEmail(ctx context.Context, email string) (*database.Contact, error) {
	if email == "" {
		return nil, fmt.Errorf("email is empty")
	}

	callCtx, cancelCallCtx := context.WithTimeout(ctx, c.timeout)
	defer cancelCallCtx()

	contact := database.Contact{}
	// find contact by email
	err := c.Pool.QueryRow(
		callCtx,
		`SELECT 
		(
			id, phoneNumber, email, linkedId, 
			linkPrecedence, createdAt, updatedAt
		)
		FROM `+TableName_Contact+`
		WHERE email = $1
		ORDER BY updatedAt DESC
		LIMIT 1`,
		email,
	).Scan(
		&contact.Id,
		&contact.PhoneNumber,
		&contact.Email,
		&contact.LinkedId,
		&contact.LinkPrecedence,
		&contact.CreatedAt,
		&contact.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("finding contact in postgres : %s", err.Error())
	}

	return &contact, nil
}

func (c Client) FindContactByPhone(ctx context.Context, phone string) (*database.Contact, error) {
	if phone == "" {
		return nil, fmt.Errorf("phone is empty")
	}

	callCtx, cancelCallCtx := context.WithTimeout(ctx, c.timeout)
	defer cancelCallCtx()

	contact := database.Contact{}
	// find contact by phone
	err := c.Pool.QueryRow(
		callCtx,
		`SELECT
		(
			id, phoneNumber, email, linkedId,
			linkPrecedence, createdAt, updatedAt
		)
		FROM `+TableName_Contact+`
		WHERE phoneNumber = $1
		ORDER BY updatedAt DESC
		LIMIT 1`,
		phone,
	).Scan(
		&contact.Id,
		&contact.PhoneNumber,
		&contact.Email,
		&contact.LinkedId,
		&contact.LinkPrecedence,
		&contact.CreatedAt,
		&contact.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("finding contact in postgres : %s", err.Error())
	}

	return &contact, nil
}

func (c Client) FindAllContacts(ctx context.Context, email, phone string) ([]*database.Contact, error) {
	callCtx, cancelCallCtx := context.WithTimeout(ctx, c.timeout)
	defer cancelCallCtx()

	rows, err := c.Pool.Query(
		callCtx,
		`WITH RECURSIVE linked_contacts AS (
			SELECT 
				id, phoneNumber, email, 
				linkedId, linkPrecedence
			FROM Contact
			WHERE phoneNumber = $1 OR email = $2
			UNION
			SELECT cc.id, cc.phoneNumber, cc.email, cc.linkedId, cc.linkPrecedence
			FROM Contact cc
			INNER JOIN linked_contacts lc ON cc.linkedId = lc.id OR lc.linkedId = cc.id
		)
		SELECT 
			id, phoneNumber, email, 
			linkedId, linkPrecedence 
		FROM linked_contacts;`,
		phone, email,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("finding all contacts in postgres : %s", err.Error())
	}
	contacts := make([]*database.Contact, 0)
	// add all contacts to slice
	for rows.Next() {
		contact := database.Contact{}
		err := rows.Scan(
			&contact.Id,
			&contact.PhoneNumber,
			&contact.Email,
			&contact.LinkedId,
			&contact.LinkPrecedence,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning contact in postgres : %s", err.Error())
		}
		contacts = append(contacts, &contact)
	}
	// close rows
	rows.Close()
	// return contacts
	return contacts, nil
}
