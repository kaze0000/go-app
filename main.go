package main

import (
	"database/sql"
	"fmt"
	"go-app/pkg/adapters"
	"go-app/pkg/adapters/middleware"
	"go-app/pkg/infra"
	"go-app/pkg/usecase"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func open(path string, count uint) *sql.DB {
	db, err := sql.Open("mysql", path)
	if err != nil {
				log.Fatal("open error:", err)
	}

	if err := db.Ping(); err != nil {
		time.Sleep(time.Second* 2)
		count --
		fmt.Printf("retry... count:%v\n", count)
		return open(path, count)
	}

	fmt.Println("db connected!!")
	return db
}

func connectDB() *sql.DB {
	var path string = fmt.Sprintf("%s:%s@tcp(db:3306)/%s?charset=utf8&parseTime=true",
		os.Getenv("MYSQL_USER"),os.Getenv("MYSQL_PASSWORD"),os.Getenv("MYSQL_DATABASE"))
	return open(path, 100)
}

func main() {
	db := connectDB()
	defer db.Close()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	e := echo.New()

	jWTMiddleware := middleware.JWTMiddleware(os.Getenv("jwtSecretKey"))
	// users
	userRepo := infra.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo, os.Getenv("jwtSecretKey"))
	userHandler := adapters.NewUserHandler(userUsecase)
	e.POST("/register", userHandler.Register)
	e.POST("/login", userHandler.Login)
	// authorized
	authorized := e.Group("")
	authorized.Use(jWTMiddleware)
	// messges
	chatGPTAPI := infra.NewChatGPTAPIClient(os.Getenv("chatGPTAPIKey"))
	messageRepo := infra.NewMessageRepository(db)
	messageUsecase := usecase.NewMessageUsecase(messageRepo, chatGPTAPI)
	messageHandler := adapters.NewMessageHandler(messageUsecase)
	authorized.POST("/messages", messageHandler.SendMessageAndSaveResponse)
	authorized.GET("/messages", messageHandler.GetMessagesAndResponseByUserID)
  // messageOwnership
	ownershipGroup := authorized.Group("")
	ownershipGroup.Use(middleware.CheckMessageOwnership(messageUsecase))
	ownershipGroup.PUT("/messages/:id", messageHandler.UpdateMessageContent)
	ownershipGroup.DELETE("/messages/:id", messageHandler.DeleteMessage)

	e.Start(":" + "8080")
}
