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
		return nil, err
	}

	restore := func() error {
		return term.Restore(int(os.Stdin.Fd()), oldState)
	}

	return restore, nil
}

func readFromWebSocket(ctx context.Context, conn *websocket.Conn, errC chan error) {
	for {
		_, message, err := conn.Read(ctx)
		if err != nil {
			status := websocket.CloseStatus(err)
			if isNormalClose(status) {
				errC <- nil
			} else {
				errC <- fmt.Errorf("read error: %v", err)
			}
			return
		}
		_, _ = os.Stdout.Write(message)
	}
}

func writeToWebSocket(ctx context.Context, conn *websocket.Conn, errC chan error) {
	buf := make([]byte, readBufferSize)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			errC <- err
			return
		}

		err = conn.Write(ctx, websocket.MessageText, buf[:n])
		if err != nil {
			errC <- err
			return
		}
	}
}

func Connect(wsURI string) error {
	ctx := context.Background()
	conn, resp, err := websocket.Dial(ctx, wsURI, nil)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("websocket dial failed: %v (HTTP %d)", err, resp.StatusCode)
		}
		return fmt.Errorf("websocket dial failed: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	restore, err := setupTerminal()
	if err != nil {
		return err
	}
	defer restore()

	errC := make(chan error, 1)

	go readFromWebSocket(ctx, conn, errC)
	go writeToWebSocket(ctx, conn, errC)

	return <-errC
}
