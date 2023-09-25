package http

import "fmt"

type CustomError struct {
	Description string
}

func (e *CustomError) Error() string {
	return e.Description
}

type HTTPTerminationError struct {
	err error
}

func (e *HTTPTerminationError) Error() string {
	return fmt.Sprintf("Shutdown HTTP server error: %v", e.err)
}

var (
	ErrSomethingWentWrong = &CustomError{Description: "Something went wrong"}
	ErrCouldNotFindItem   = &CustomError{Description: "Could not find the item"}
)
