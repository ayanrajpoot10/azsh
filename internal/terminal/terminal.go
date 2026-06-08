package terminal

import (
	"context"
	"fmt"
	"os"

	"github.com/coder/websocket"
	"golang.org/x/term"
)

const readBufferSize = 1024

func isNormalClose(status websocket.StatusCode) bool {
	return status == websocket.StatusNormalClosure ||
		status == websocket.StatusGoingAway ||
		status == -1
}

func setupTerminal() (func() error, error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("raw terminal: %w", err)
	}

	restore := func() error {
		return term.Restore(int(os.Stdin.Fd()), oldState)
	}

	return restore, nil
}

func readFromWebSocket(ctx context.Context, conn *websocket.Conn, errCh chan error) {
	for {
		_, message, err := conn.Read(ctx)
		if err != nil {
			status := websocket.CloseStatus(err)
			if isNormalClose(status) {
				errCh <- nil
			} else {
				errCh <- fmt.Errorf("read error: %w", err)
			}
			return
		}
		
		if len(message) == 0 {
			errCh <- nil
			return
		}

		_, _ = os.Stdout.Write(message)
	}
}

func writeToWebSocket(ctx context.Context, conn *websocket.Conn, errCh chan error) {
	buf := make([]byte, readBufferSize)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			errCh <- fmt.Errorf("stdin read: %w", err)
			return
		}

		err = conn.Write(ctx, websocket.MessageText, buf[:n])
		if err != nil {
			errCh <- fmt.Errorf("ws write: %w", err)
			return
		}
	}
}

func Connect(wsURI string) error {
	ctx := context.Background()
	conn, resp, err := websocket.Dial(ctx, wsURI, nil)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("websocket dial: %w (HTTP %d)", err, resp.StatusCode)
		}
		return fmt.Errorf("websocket dial: %w", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	restore, err := setupTerminal()
	if err != nil {
		return err
	}
	defer restore()

	errCh := make(chan error, 1)

	go readFromWebSocket(ctx, conn, errCh)
	go writeToWebSocket(ctx, conn, errCh)

	return <-errCh
}
