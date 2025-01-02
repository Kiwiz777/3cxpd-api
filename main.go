package main

import (
	"log"
	"time"

	"github.com/kiwiz777/3cxpd-api/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)


type NewStorageClient struct {
	DB *database.StorageClient
}

func NewDatabaseHandler() (*NewStorageClient, error) {
	db, err := database.NewStorageClient()
	if err != nil {
		return nil, err
	}
	return &NewStorageClient{
		DB: db,
	}, nil
}

func main() {
	handler, err := NewDatabaseHandler()
	if err != nil {
		log.Fatalf("Failed to create job handler: %v", err)
	}

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*", "http://localhost:5173", "https://aura.mcn.red", "http://aura.mcn.red"},
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	}))
	router.POST("/login", handler.Login)
	router.POST("/check-token", handler.CheckToken)
	//user
	// router.GET("/users", handler.CheckAuth, handler.GetUsers)
	// router.POST("/user", handler.CheckAuth, handler.CreateUser)
	// router.GET("/user/:id", handler.CheckAuth, handler.GetUser)

	router.POST("/contact", handler.CheckAuth, handler.AddContact)
	router.POST("/contacts", handler.CheckAuth, handler.BulkAddContacts)
	router.PUT("/contact", handler.CheckAuth, handler.UpdateContact)
	router.GET("/contacts", handler.CheckAuth, handler.GetContacts)
	router.GET("/pending", handler.CheckAuth, handler.GetContactsPending)
	router.GET("/next", handler.CheckAuth, handler.NextContact)


	router.Run(":3000")
}
