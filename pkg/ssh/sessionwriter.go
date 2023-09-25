package ssh

import (
	"bytes"
	"fmt"
	"io"

	"github.com/AlexEkdahl/gotit/pkg/util"
)

type SessionWriter struct {
	w io.Writer
}

func NewSessionWriter(w io.Writer) *SessionWriter {
	return &SessionWriter{
		w: w,
	}
}

func (sw *SessionWriter) write(msg string) error {
	_, err := io.WriteString(sw.w, msg)
	if err != nil {
		return fmt.Errorf("Error writing to session: %v", err)
	}
	return nil
}

func (sw *SessionWriter) WriteWelcomeMsg(user string) {
	var buf bytes.Buffer

	buf.WriteString("\033c") // Clear screen

	buf.WriteString(colored("📫  gotit.sh verified user\n", util.Gray))
	buf.WriteString("\n")

	buf.WriteString(colored(fmt.Sprintf("Welcome %s!", user), util.SoftYellow))
	buf.WriteString("\n")

	buf.WriteString(colored("Your connection stays open until someone downloads your file.", util.SoftGreen))
	buf.WriteString("\n")

	_ = sw.write(buf.String())
}

// Helper function to write a string in a certain color
func colored(s, color string) string {
	return fmt.Sprintf("%s%s%s", color, s, util.Reset)
}

func (sw *SessionWriter) WriteTypeUsage() {
	usage := "ssh gotit.sh [MIMETYPE] < response.json"
	_ = sw.write(usage)
}

func (sw *SessionWriter) WriteTransferDone(speed float64) {
	var buf bytes.Buffer
	buf.WriteString("\n")
	buf.WriteString(colored("Data transfered with no errors.\n", util.SoftGreen))
	buf.WriteString(colored("Transfer speed: ", util.SoftYellow))
	buf.WriteString(colored(fmt.Sprintf("%0.f Mb/s\n", speed), util.SoftYellow))

	_ = sw.write(buf.String())
}

func (sw *SessionWriter) WriteError(err error) {
	_ = sw.write(fmt.Sprintf("%sError: %s%s", util.Red, err, util.Reset))
}

func (sw *SessionWriter) WriteURL(addr string) {
	var buf bytes.Buffer
	buf.WriteString("\n")
	buf.WriteString(colored("Share link: ", util.SoftGreen))
	buf.WriteString("\n")
	buf.WriteString(colored(addr, util.LinkColor))
	buf.WriteString("\n")
	buf.WriteString("\n")
	buf.WriteString(colored("Direct link: ", util.SoftGreen))
	buf.WriteString("\n")
	buf.WriteString(colored(addr, util.LinkColor))
	buf.WriteString("\n")

	_ = sw.write(buf.String())
}
