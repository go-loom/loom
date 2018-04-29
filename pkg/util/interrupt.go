package util

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
)

func Interrupt(cancel <-chan struct{}) error {
	c := make(chan os.Singal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-c:
		return fmt.Errorf("received signal %s", sig)
	case <-cancel:
		return errors.New("canceled")
	}
}
