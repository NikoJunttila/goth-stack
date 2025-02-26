package helloworld

import (
	"fmt"
	"gothstack/app/db"
	"time"

	"gorm.io/gorm"
)

// Event name constants
const (
	HelloworldNewEvent = "helloworld.newEvent"
)

type HelloWorld struct {
	Message string
}

type HelloworldMessage struct {
	gorm.Model

	Message   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func createMessage(message string) (HelloworldMessage, error) {
	fmt.Println(message)
	hello := HelloworldMessage{
		Message: "Hello world!",
	}
	result := db.Get().Create(&hello)
	return hello, result.Error
}
func listMessages() ([]HelloworldMessage, error) {
	var messages []HelloworldMessage
	result := db.Get().Order("created_at desc").Find(&messages)
	return messages, result.Error
}
