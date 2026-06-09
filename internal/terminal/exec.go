package terminal

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/coder/websocket"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\].*?\x1b\\`)

func ExecCommand(wsURI, command string) error {
	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, wsURI, nil)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}

	readySentinel := fmt.Sprintf("__AZSH_READY_%x__", rand.Int63())

	init := fmt.Sprintf("PS1='' && stty -echo\necho %s\n", readySentinel)
	if err := conn.Write(ctx, websocket.MessageText, []byte(init)); err != nil {
		conn.Close(websocket.StatusNormalClosure, "")
		return fmt.Errorf("write init: %w", err)
	}

	var buf strings.Builder
	for {
		_, msg, err := conn.Read(ctx)
		if err != nil {
			conn.Close(websocket.StatusNormalClosure, "")
			return fmt.Errorf("read init: %w", err)
		}
		buf.Write(msg)
		if strings.Contains(buf.String(), readySentinel) {
			break
		}
	}

	doneSentinel := fmt.Sprintf("__AZSH_DONE_%x__", rand.Int63())
	fullCmd := fmt.Sprintf("%s\necho %s\n", command, doneSentinel)
	if err := conn.Write(ctx, websocket.MessageText, []byte(fullCmd)); err != nil {
		conn.Close(websocket.StatusNormalClosure, "")
		return fmt.Errorf("write command: %w", err)
	}

	buf.Reset()
	var out strings.Builder
	for {
		_, msg, err := conn.Read(ctx)
		if err != nil {
			break
		}
		buf.Write(msg)
		if idx := strings.Index(buf.String(), doneSentinel); idx >= 0 {
			out.WriteString(buf.String()[:idx])
			break
		}
	}

	result := cleanOutput(out.String())
	if result != "" {
		fmt.Println(result)
	}
	go conn.Close(websocket.StatusNormalClosure, "")
	return nil
}

func cleanOutput(s string) string {
	s = ansiRe.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "")
	return strings.TrimSpace(s)
}
