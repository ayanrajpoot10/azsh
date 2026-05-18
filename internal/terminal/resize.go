package terminal

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

func HandleResize(onResize func(width, height int)) {
	sigWinch := make(chan os.Signal, 1)
	signal.Notify(sigWinch, syscall.SIGWINCH)

	go func() {
		for range sigWinch {
			w, h, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil {
				onResize(w, h)
			}
		}
	}()
}
