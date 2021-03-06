package commands

import (
	"errors"
	"os"
	"time"

	"git.sr.ht/~rjarry/aerc/widgets"
)

type PrintWorkDir struct{}

func init() {
	register(PrintWorkDir{})
}

func (PrintWorkDir) Aliases() []string {
	return []string{"pwd"}
}

func (PrintWorkDir) Complete(aerc *widgets.Aerc, args []string) []string {
	return nil
}

func (PrintWorkDir) Execute(aerc *widgets.Aerc, args []string) error {
	if len(args) != 1 {
		return errors.New("Usage: pwd")
	}
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	aerc.PushStatus(pwd, 10*time.Second)
	return nil
}
