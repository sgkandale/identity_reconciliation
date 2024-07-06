package server

import (
	"log"
	"net/http"

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

}
