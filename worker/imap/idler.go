package imap

import (
	"fmt"
	"sync"
	"time"

	"git.sr.ht/~rjarry/aerc/logging"
	"git.sr.ht/~rjarry/aerc/worker/types"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

var (
	errIdleTimeout   = fmt.Errorf("idle timeout")
	errIdleModeHangs = fmt.Errorf("idle mode hangs; waiting to reconnect")
)

// idler manages the idle mode of the imap server. Enter idle mode if there's
// no other task and leave idle mode when a new task arrives. Idle mode is only
// used when the client is ready and connected. After a connection loss, make
// sure that idling returns gracefully and the worker remains responsive.
type idler struct {
	sync.Mutex
	config  imapConfig
	client  *imapClient
	worker  *types.Worker
	last    time.Time
	stop    chan struct{}
	done    chan error
	waiting bool
	idleing bool
}

func newIdler(cfg imapConfig, w *types.Worker) *idler {
	return &idler{config: cfg, worker: w, done: make(chan error)}
}

func (i *idler) SetClient(c *imapClient) {
	i.Lock()
	i.client = c
	i.Unlock()
}

func (i *idler) setWaiting(wait bool) {
	i.Lock()
	i.waiting = wait
	i.Unlock()
}

func (i *idler) isWaiting() bool {
	i.Lock()
	defer i.Unlock()
	return i.waiting
}

func (i *idler) isReady() bool {
	i.Lock()
	defer i.Unlock()
	return (!i.waiting && i.client != nil &&
		i.client.State() == imap.SelectedState)
}

func (i *idler) Start() {
	if i.isReady() {
		i.stop = make(chan struct{})

		go func() {
			defer logging.PanicHandler()
			select {
			case <-i.stop:
				// debounce idle
				i.log("=>(idle) [debounce]")
				i.done <- nil
			case <-time.After(i.config.idle_debounce):
				// enter idle mode
				i.idleing = true
				i.log("=>(idle)")
				now := time.Now()
				err := i.client.Idle(i.stop,
					&client.IdleOptions{
						LogoutTimeout: 0,
						PollInterval:  0,
					})
				i.idleing = false
				i.done <- err
				i.log("elapsed idle time:",
					time.Since(now))
			}
		}()

	} else if i.isWaiting() {
		i.log("not started: wait for idle to exit")
	} else {
		i.log("not started: client not ready")
	}
}

func (i *idler) Stop() error {
	var reterr error
	if i.isReady() {
		close(i.stop)
		select {
		case err := <-i.done:
			if err == nil {
				i.log("<=(idle)")
			} else {
				i.log("<=(idle) with err:", err)
			}
			reterr = nil
		case <-time.After(i.config.idle_timeout):
			i.log("idle err (timeout); waiting in background")

			i.log("disconnect done->")
			i.worker.PostMessage(&types.Done{
				Message: types.RespondTo(&types.Disconnect{}),
			}, nil)

			i.waitOnIdle()

			reterr = errIdleTimeout
		}
	} else if i.isWaiting() {
		i.log("not stopped: still idleing/hanging")
		reterr = errIdleModeHangs
	} else {
		i.log("not stopped: client not ready")
		reterr = nil
	}
	return reterr
}

func (i *idler) waitOnIdle() {
	i.setWaiting(true)
	i.log("wait for idle in background")
	go func() {
		defer logging.PanicHandler()
		select {
		case err := <-i.done:
			if err == nil {
				i.log("<=(idle) waited")
				i.log("connect done->")
				i.worker.PostMessage(&types.Done{
					Message: types.RespondTo(&types.Connect{}),
				}, nil)
			} else {
				i.log("<=(idle) waited; with err:", err)
			}
			i.setWaiting(false)
			i.stop = make(chan struct{})
			i.log("restart")
			i.Start()
			return
		}
	}()
}

func (i *idler) log(args ...interface{}) {
	header := fmt.Sprintf("idler (%p) [idle:%t,wait:%t]", i, i.idleing, i.waiting)
	i.worker.Logger.Println(append([]interface{}{header}, args...)...)
}
