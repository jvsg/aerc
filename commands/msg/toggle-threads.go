package msg

import (
	"errors"

	"git.sr.ht/~rjarry/aerc/lib/statusline"
	"git.sr.ht/~rjarry/aerc/widgets"
)

type ToggleThreads struct{}

func init() {
	register(ToggleThreads{})
}

func (ToggleThreads) Aliases() []string {
	return []string{"toggle-threads"}
}

func (ToggleThreads) Complete(aerc *widgets.Aerc, args []string) []string {
	return nil
}

func (ToggleThreads) Execute(aerc *widgets.Aerc, args []string) error {
	if len(args) != 1 {
		return errors.New("Usage: toggle-threads")
	}
	h := newHelper(aerc)
	acct, err := h.account()
	if err != nil {
		return err
	}
	store, err := h.store()
	if err != nil {
		return err
	}
	defer store.Reselect(store.Selected())
	store.SetBuildThreads(!store.BuildThreads())
	acct.SetStatus(statusline.Threading(store.BuildThreads()))
	acct.Messages().Invalidate()
	return nil
}
