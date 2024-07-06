package server

import (
	"identity/database"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (s *Server) RegisterHandlers() {
	s.POST("/identify", s.Identify)
}

func (s *Server) Identify(ctx *gin.Context) {
	var reqBody IdentifyRequest
	err := ctx.BindJSON(&reqBody)
	if err != nil {
		log.Print("[ERROR] reading request body in server.Identify : ", err.Error())
		ctx.JSON(http.StatusBadRequest, GetGeneralResponseError(err))
		return
	}

	// check for existing contacts
	existingRecentContact, err := s.DbConn.FindRecentContact(
		ctx.Request.Context(),
		&database.Contact{
			Email:       reqBody.Email,
			PhoneNumber: reqBody.PhoneNumber,
		},
	)
	if err != nil {
		log.Print("[ERROR] finding existing contact in server.Identify : ", err.Error())
		ctx.JSON(http.StatusInternalServerError, GetGeneralResponseError(err))
		return
	}

	// no contact matched with the given email or phone number
	// create a new contact
	if existingRecentContact == nil {
		newContact := &database.Contact{
			PhoneNumber:    reqBody.PhoneNumber,
			Email:          reqBody.Email,
			LinkPrecedence: database.ContactLinkPrecedence_Primary,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// insert in db
		err = s.DbConn.PutContact(
			ctx.Request.Context(),
			newContact,
		)
		if err != nil {
			log.Print("[ERROR] inserting new contact in server.Identify : ", err.Error())
			ctx.JSON(http.StatusInternalServerError, GetGeneralResponseError(err))
			return
		}

		// return response
		ctx.JSON(
			http.StatusOK,
			IdentifyResponse{
				Contact: IdentifyResponseContact{
					// PrimaryContactId: 0, // pending
					Emails:              []string{newContact.Email},
					PhoneNumbers:        []string{newContact.PhoneNumber},
					SecondaryContactIds: []int64{},
				},
			},
		)
	}

	// if recent contact exists, find all contacts
	allContacts, err := s.DbConn.FindAllContacts(
		ctx.Request.Context(),
		reqBody.Email,
		reqBody.PhoneNumber,
	)
	if err != nil {
		log.Print("[ERROR] finding all contacts in server.Identify : ", err.Error())
		ctx.JSON(http.StatusInternalServerError, GetGeneralResponseError(err))
		return
	}

	// prepare a formatted data
	primaryContactId := int64(0)
	primaryEmail := ""
	secondaryEmails := []string{}
	primaryPhone := ""
	secondaryPhones := []string{}
	secondaryIds := []int64{}

	for _, eachContact := range allContacts {
		if eachContact.LinkPrecedence == database.ContactLinkPrecedence_Primary {
			primaryContactId = eachContact.Id
			primaryEmail = eachContact.Email
			primaryPhone = eachContact.PhoneNumber
		} else {
			secondaryIds = append(secondaryIds, eachContact.Id)
			if eachContact.Email != "" {
				secondaryEmails = append(secondaryEmails, eachContact.Email)
			}
			if eachContact.PhoneNumber != "" {
				secondaryPhones = append(secondaryPhones, eachContact.PhoneNumber)
			}
		}
	}

	ctx.JSON(
		http.StatusOK,
		IdentifyResponse{
			Contact: IdentifyResponseContact{
				PrimaryContactId:    primaryContactId,
				Emails:              append([]string{primaryEmail}, secondaryEmails...),
				PhoneNumbers:        append([]string{primaryPhone}, secondaryPhones...),
				SecondaryContactIds: secondaryIds,
			},
		},
	)
}
