package postgres

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

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
	if contact.LinkedId != nil && *contact.LinkedId != 0 {
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

	// if both are present, and different
	if lastContactByPhone != nil && lastContactByEmail != nil {
		// if both are different
		// set earliest one to primary, and rest all to secondary,
		if lastContactByPhone.Id != lastContactByEmail.Id {
			// get earliest one
			allContacts, err := c.FindAllContacts(ctx, childContact.Email, childContact.PhoneNumber)
			if err != nil {
				return nil, fmt.Errorf("finding all contacts for linking update : %s", err.Error())
			}
			// sort by created at
			sort.SliceStable(allContacts, func(i, j int) bool {
				return allContacts[i].CreatedAt.Before(allContacts[j].CreatedAt)
			})
			// update all others after first without linkedid to first one's id
			baseId := allContacts[0].Id
			linkIdUpdateWg := sync.WaitGroup{}
			for i := 1; i < len(allContacts); i++ {
				linkIdUpdateWg.Add(1)
				go func(i int) {
					defer linkIdUpdateWg.Done()
					err := c.UpdateLinkId(ctx, allContacts[i].Id, baseId)
					if err != nil {
						fmt.Println("error updating link id : ", err.Error())
					}
				}(i)
			}
			linkIdUpdateWg.Wait()
		}

		// and return the latest one
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
			id, COALESCE(phoneNumber, ''), 
			COALESCE(email, ''), linkedId, 
			linkPrecedence, createdAt, updatedAt
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
			id, COALESCE(phoneNumber, ''), 
			COALESCE(email, ''), linkedId,
			linkPrecedence, createdAt, updatedAt
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
			id, COALESCE(phoneNumber, ''), 
			COALESCE(email, ''),
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

func (c *Client) UpdateLinkId(ctx context.Context, contactId, linkedId int64) error {
	callCtx, cancelCallCtx := context.WithTimeout(ctx, c.timeout)
	defer cancelCallCtx()

	_, err := c.Pool.Exec(
		callCtx,
		`UPDATE `+TableName_Contact+`
		SET 
			linkedId = $1,
			updatedAt = $2,
			linkPrecedence = $3
		WHERE id = $4`,
		linkedId,
		time.Now(),
		database.ContactLinkPrecedence_Secondary,
		contactId,
	)
	if err != nil {
		return fmt.Errorf("updating link id in postgres : %s", err.Error())
	}

	return nil
}
