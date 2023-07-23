package server

import (
	"fmt"
	"io"

	"github.com/AlexEkdahl/gotit/utils/colors"
)

type sessionWriter struct {
	w io.Writer
}

type SessionWriter interface {
	WriteWelcomeMsg(user string) error
	WriteTypeUsage() error
	WriteTransferSpeed(speed float64) error
	WriteError(err error) error
	WriteURL(addr string) error
}

func NewSessionWriter(w io.Writer) SessionWriter {
	return &sessionWriter{
		w: w,
	}
}

func (sw *sessionWriter) write(msg string) error {
	_, err := io.WriteString(sw.w, msg)
	if err != nil {
		return fmt.Errorf("Error writing to session: %v", err)
	}
	return nil
}

func (sw *sessionWriter) WriteWelcomeMsg(user string) error {
	return sw.write(fmt.Sprintf("%sWelcome, %s!%s\n", colors.Green, user, colors.Reset))
}

func (sw *sessionWriter) WriteTypeUsage() error {
	usage := "ssh gotit.sh [MIMETYPE] < response.json"
	return sw.write(usage)
}

func (sw *sessionWriter) WriteTransferSpeed(speed float64) error {
	return sw.write(fmt.Sprintf("Transfer speed: %.2f Mb/s\n", speed))
}

func (sw *sessionWriter) WriteError(err error) error {
	return sw.write(fmt.Sprintf("%sError: %s%s", colors.Red, err, colors.Reset))
}

func (sw *sessionWriter) WriteURL(addr string) error {
	return sw.write(addr)
}
