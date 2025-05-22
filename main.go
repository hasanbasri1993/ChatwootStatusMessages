package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

type Message struct {
	ID     uint `gorm:"primaryKey"`
	Status int  `gorm:"column:status"` // Assuming the status field in your DB is named "status"
}

// MessageResponse New field for the response, not stored in the DB
type MessageResponse struct {
	ID     uint   `json:"ID"`
	Status string `json:"Status"`
}

func ConnectDB() *gorm.DB {
	//dsn := "host=xxxxxx user=xxxxxx password=xxxxxx dbname=xxxxxx port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	dsn, errDSN := checkEnv("DSN_DATABASE")
	if errDSN != nil {
		panic("failed message: " + errDSN.Error())
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	fmt.Println("Database connection established")
	if err != nil {
		panic("Failed to connect to database!")
	}
	return db
}

type RequestBody struct {
	IDs []int `json:"ids"`
}

func (m *Message) GetStatusText() string {
	switch m.Status {
	case 2:
		return "Read"
	case 1:
		return "Delivery"
	case 0:
		return "Sent"
	default:
		return "Unknown"
	}
}

func getMessages(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Initialize your struct to hold the POST request body
		var reqBody RequestBody

		// Parse the request body into your struct
		if err := c.BodyParser(&reqBody); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Bad request"})
		}

		// Now you can use reqBody.IDs just like you used the ids from the query string before
		var messages []Message
		result := db.Where("id IN ?", reqBody.IDs).Find(&messages)

		if result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve messages"})
		}

		// Convert messages to include text status instead of numeric
		messagesResponse := make([]MessageResponse, len(messages))
		for i, msg := range messages {
			messagesResponse[i] = MessageResponse{
				ID:     msg.ID,
				Status: msg.GetStatusText(), // Use the status text instead of the numeric value
			}
		}

		return c.JSON(messagesResponse)
	}
}

func checkEnv(env string) (string, error) {
	check := os.Getenv(env)
	if check == "" {
		return "", fmt.Errorf("not found environment of %s", env)
	}
	return check, nil
}

func setupRoutes(app *fiber.App, db *gorm.DB) {
	app.Post("/messages", getMessages(db))
}

func main() {
	db := ConnectDB()
	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Fiber",
		AppName:       "Chatwoot Messages Status v1.0.2",
	})
	setupRoutes(app, db)
	err := app.Listen(":3003")
	if err != nil {
		return
	}
}
