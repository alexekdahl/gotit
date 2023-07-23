package server

import "fmt"

type CustomError struct {
	Description string
}

func (e *CustomError) Error() string {
	return e.Description
}

type CopyDataError struct {
	err error
}

func (e *CopyDataError) Error() string {
	return fmt.Sprintf("Error copying data: %v", e.err)
}

type FileError struct {
	err error
}

func (e *FileError) Error() string {
	return fmt.Sprintf("Error determining file extension: %v", e.err)
}

type ReadDataError struct {
	err error
}

func (e *ReadDataError) Error() string {
	return fmt.Sprintf("Error reading data: %v", e.err)
}

type FileExtensionError struct {
	err error
}

func (e *FileExtensionError) Error() string {
	return fmt.Sprintf("Error determining file extension: %v", e.err)
}

type TunnelWriteError struct {
	err error
}

func (e *TunnelWriteError) Error() string {
	return fmt.Sprintf("Error writing data to tunnel: %v", e.err)
}

type SSHTerminationError struct {
	err error
}

func (e *SSHTerminationError) Error() string {
	return fmt.Sprintf("Error closing SSH server: %v", e.err)
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
