package terminal

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/coder/websocket"
	"golang.org/x/term"
)

func Connect(wsURI string) error {
	ctx := context.Background()
	header := http.Header{}
	header.Set("Origin", "https://ux.console.azure.com")
	header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36")

	c, resp, err := websocket.Dial(ctx, wsURI, &websocket.DialOptions{
		HTTPHeader: header,
	})
	if err != nil {
		if resp != nil {
			return fmt.Errorf("dial err: %v - HTTP status: %d", err, resp.StatusCode)
		}
		return fmt.Errorf("dial err: %v", err)
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	errC := make(chan error, 1)

	go func() {
		for {
			_, message, err := c.Read(ctx)
			if err != nil {
				status := websocket.CloseStatus(err)
				if status == websocket.StatusNormalClosure || status == websocket.StatusGoingAway || status == -1 {
					errC <- nil // Normal close or context cancelled/EOF
				} else {
					errC <- fmt.Errorf("read error: %v", err)
				}
				return
			}
			_, _ = os.Stdout.Write(message)
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				errC <- err
				return
			}
			err = c.Write(ctx, websocket.MessageText, buf[:n])
			if err != nil {
				errC <- err
				return
			}
		}
	}()

	return <-errC
}
