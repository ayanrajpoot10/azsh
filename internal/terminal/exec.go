package terminal

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/coder/websocket"
)

var ansiRe = regexp.MustCompile(`\x1b\[[\x30-\x3f]*[\x20-\x2f]*[\x40-\x7e]|\x1b\].*?\x1b\\`)

func ExecCommand(wsURI, command string) error {
	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, wsURI, nil)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}
	sentinel := fmt.Sprintf("__AZSH_EXEC_%x__", rand.Int63())

	init := fmt.Sprintf("PS1='%s'; stty -echo\n", sentinel)
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
		if strings.Count(buf.String(), sentinel) >= 2 {
			break
		}
	}

	if err := conn.Write(ctx, websocket.MessageText, []byte(command+"\n")); err != nil {
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
		if idx := strings.Index(buf.String(), sentinel); idx >= 0 {
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
