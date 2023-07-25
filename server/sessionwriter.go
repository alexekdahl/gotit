package server

import (
	"bytes"
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
	WriteTransferDone(speed float64) error
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
	var buf bytes.Buffer

	buf.WriteString("\033c") // Clear screen

	buf.WriteString(colored("📫  gotit.sh verified user\n", colors.Gray))
	buf.WriteString("\n")

	buf.WriteString(colored(fmt.Sprintf("Welcome %s!", user), colors.SoftYellow))
	buf.WriteString("\n")

	buf.WriteString(colored("Your connection stays open until someone downloads your file.", colors.SoftGreen))
	buf.WriteString("\n")

	return sw.write(buf.String())
}

// Helper function to write a string in a certain color
func colored(s, color string) string {
	return fmt.Sprintf("%s%s%s", color, s, colors.Reset)
}

func (sw *sessionWriter) WriteTypeUsage() error {
	usage := "ssh gotit.sh [MIMETYPE] < response.json"
	return sw.write(usage)
}

func (sw *sessionWriter) WriteTransferDone(speed float64) error {
	var buf bytes.Buffer
	buf.WriteString("\n")
	buf.WriteString(colored("Data transfered with no errors.\n", colors.SoftGreen))
	buf.WriteString(colored("Transfer speed: ", colors.SoftYellow))
	buf.WriteString(colored(fmt.Sprintf("%0.f Mb/s\n", speed), colors.SoftYellow))

	return sw.write(buf.String())
}

func (sw *sessionWriter) WriteError(err error) error {
	return sw.write(fmt.Sprintf("%sError: %s%s", colors.Red, err, colors.Reset))
}

func (sw *sessionWriter) WriteURL(addr string) error {
	var buf bytes.Buffer
	buf.WriteString("\n")
	buf.WriteString(colored("Share link: ", colors.SoftGreen))
	buf.WriteString("\n")
	buf.WriteString(colored(addr, colors.LinkColor))
	buf.WriteString("\n")
	buf.WriteString("\n")
	buf.WriteString(colored("Direct link: ", colors.SoftGreen))
	buf.WriteString("\n")
	buf.WriteString(colored(addr, colors.LinkColor))
	buf.WriteString("\n")

	return sw.write(buf.String())
}
