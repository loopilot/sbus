package sbus

import "fmt"

type Error struct {
	Message     string
	Code        int16
	Description string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Description)
}

var (
	ErrInvalidHandler    = &Error{"INVALID_HANDLER", 1, "The handler is not compatible with the topic"}
	ErrTopicDoesNotExist = &Error{"TOPIC_DOES_NOT_EXIST", 2, "The topic does not exist"}
	ErrBusDoesNotExist   = &Error{"BUS_DOES_NOT_EXIST", 3, "The bus does not exist"}
)
