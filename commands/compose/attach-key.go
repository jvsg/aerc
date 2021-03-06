package compose

import (
	"errors"

	"git.sr.ht/~rjarry/aerc/widgets"
)

type AttachKey struct{}

func init() {
	register(AttachKey{})
}

func (AttachKey) Aliases() []string {
	return []string{"attach-key"}
}

func (AttachKey) Complete(aerc *widgets.Aerc, args []string) []string {
	return nil
}

func (AttachKey) Execute(aerc *widgets.Aerc, args []string) error {
	if len(args) != 1 {
		return errors.New("Usage: attach-key")
	}

	composer, _ := aerc.SelectedTab().(*widgets.Composer)

	composer.SetAttachKey(!composer.AttachKey())
	return nil
}
