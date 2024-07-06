package server

import (
	"context"
	"fmt"
	"identity/config"
	"identity/database"
	"identity/database/postgres"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

type Server struct {
	*gin.Engine

	// db connection
	DbConn database.Database

	// internal variables
	port int
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func New(ctx context.Context, cfg *config.Config) *Server {
	if cfg == nil {
		log.Fatal("[ERROR] server config is nil")
	}
	if cfg.Server.Port <= 0 {
		log.Fatal("[ERROR] server port is invalid")
	}

	var dbConn database.Database
	// check database support
	switch strings.ToLower(cfg.Database.Type) {
	case "postgres":
		dbConn = postgres.New(ctx, &cfg.Database)
	default:
		log.Fatal("[ERROR] database type unsupported : ", cfg.Database.Type)
	}

	log.Print("[INFO] creating server instance")
	return &Server{
		Engine: gin.Default(),
		DbConn: dbConn,
		port:   cfg.Server.Port,
	}
}

// Serve function will start the http server.
// This is a blocking function.
func (s Server) Serve() {
	log.Print("[INFO] starting server on port : ", s.port)
	err := s.Run(fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatal("[ERROR] failed to start server : ", err)
	}
}
