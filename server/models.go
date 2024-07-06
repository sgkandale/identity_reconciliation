package server

type GeneralResponse struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

func GetGeneralResponseError(err error) GeneralResponse {
	return GeneralResponse{
		Status:  "error",
		Message: err.Error(),
	}
}

type IdentifyRequest struct {
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
}

type IdentifyResponseContact struct {
	PrimaryContactId    int64    `json:"primaryContatctId"`
	Emails              []string `json:"emails"`              // first element being email of primary contact
	PhoneNumbers        []string `json:"phoneNumbers"`        // first element being phoneNumber of primary contact
	SecondaryContactIds []int64  `json:"secondaryContactIds"` // Array of all Contact IDs that are "secondary"
}

type IdentifyResponse struct {
	Contact IdentifyResponseContact `json:"contact"`
}
