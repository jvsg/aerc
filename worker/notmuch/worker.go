//go:build notmuch
// +build notmuch

package notmuch

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"git.sr.ht/~rjarry/aerc/config"
	"git.sr.ht/~rjarry/aerc/lib/uidstore"
	"git.sr.ht/~rjarry/aerc/logging"
	"git.sr.ht/~rjarry/aerc/models"
	"git.sr.ht/~rjarry/aerc/worker/handlers"
	"git.sr.ht/~rjarry/aerc/worker/lib"
	notmuch "git.sr.ht/~rjarry/aerc/worker/notmuch/lib"
	"git.sr.ht/~rjarry/aerc/worker/types"
	"github.com/mitchellh/go-homedir"
)

func init() {
	handlers.RegisterWorkerFactory("notmuch", NewWorker)
}

var errUnsupported = fmt.Errorf("unsupported command")

const backgroundRefreshDelay = 1 * time.Minute

type worker struct {
	w                   *types.Worker
	nmEvents            chan eventType
	query               string
	currentQueryName    string
	uidStore            *uidstore.Store
	queryMapOrder       []string
	nameQueryMap        map[string]string
	db                  *notmuch.DB
	setupErr            error
	currentSortCriteria []*types.SortCriterion
}

// NewWorker creates a new notmuch worker with the provided worker.
func NewWorker(w *types.Worker) (types.Backend, error) {
	events := make(chan eventType, 20)
	return &worker{w: w,
		nmEvents: events}, nil
}

// Run starts the worker's message handling loop.
func (w *worker) Run() {
	for {
		select {
		case action := <-w.w.Actions:
			msg := w.w.ProcessAction(action)
			if err := w.handleMessage(msg); err == errUnsupported {
				w.w.PostMessage(&types.Unsupported{
					Message: types.RespondTo(msg),
				}, nil)
				w.w.Logger.Printf("ProcessAction(%T) unsupported: %v", msg, err)
			} else if err != nil {
				w.w.PostMessage(&types.Error{
					Message: types.RespondTo(msg),
					Error:   err,
				}, nil)
				w.w.Logger.Printf("ProcessAction(%T) failure: %v", msg, err)
			}
		case nmEvent := <-w.nmEvents:
			err := w.handleNotmuchEvent(nmEvent)
			if err != nil {
				w.w.Logger.Printf("notmuch event failure: %v", err)
			}
		}
	}
}

func (w *worker) done(msg types.WorkerMessage) {
	w.w.PostMessage(&types.Done{types.RespondTo(msg)}, nil)
}

func (w *worker) err(msg types.WorkerMessage, err error) {
	w.w.PostMessage(&types.Error{
		Message: types.RespondTo(msg),
		Error:   err,
	}, nil)
}
func (w *worker) handleMessage(msg types.WorkerMessage) error {
	if w.setupErr != nil {
		// only configure can recover from a config error, bail for everything else
		_, isConfigure := msg.(*types.Configure)
		if !isConfigure {
			return w.setupErr
		}
	}

	switch msg := msg.(type) {
	case *types.Unsupported:
		// No-op
	case *types.Configure:
		return w.handleConfigure(msg)
	case *types.Connect:
		return w.handleConnect(msg)
	case *types.ListDirectories:
		return w.handleListDirectories(msg)
	case *types.OpenDirectory:
		return w.handleOpenDirectory(msg)
	case *types.FetchDirectoryContents:
		return w.handleFetchDirectoryContents(msg)
	case *types.FetchDirectoryThreaded:
		return w.handleFetchDirectoryThreaded(msg)
	case *types.FetchMessageHeaders:
		return w.handleFetchMessageHeaders(msg)
	case *types.FetchMessageBodyPart:
		return w.handleFetchMessageBodyPart(msg)
	case *types.FetchFullMessages:
		return w.handleFetchFullMessages(msg)
	case *types.FlagMessages:
		return w.handleFlagMessages(msg)
	case *types.AnsweredMessages:
		return w.handleAnsweredMessages(msg)
	case *types.SearchDirectory:
		return w.handleSearchDirectory(msg)
	case *types.ModifyLabels:
		return w.handleModifyLabels(msg)
	case *types.CheckMail:
		go w.handleCheckMail(msg)
		return nil

		// not implemented, they are generally not used
		// in a notmuch based workflow
		// case *types.DeleteMessages:
		// case *types.CopyMessages:
		// 	return w.handleCopyMessages(msg)
		// case *types.AppendMessage:
		// 	return w.handleAppendMessage(msg)
		// case *types.CreateDirectory:
		// 	return w.handleCreateDirectory(msg)
		// case *types.RemoveDirectory:
		//  return w.handleRemoveDirectory(msg)
	}
	return errUnsupported
}

func (w *worker) handleConfigure(msg *types.Configure) error {
	var err error
	defer func() {
		if err == nil {
			w.setupErr = nil
			return
		}
		w.setupErr = fmt.Errorf("notmuch: %v", err)
	}()

	u, err := url.Parse(msg.Config.Source)
	if err != nil {
		w.w.Logger.Printf("error configuring notmuch worker: %v", err)
		return err
	}
	home, err := homedir.Expand(u.Hostname())
	if err != nil {
		return fmt.Errorf("could not resolve home directory: %v", err)
	}
	pathToDB := filepath.Join(home, u.Path)
	err = w.loadQueryMap(msg.Config)
	if err != nil {
		return fmt.Errorf("could not load query map configuration: %v", err)
	}
	excludedTags := w.loadExcludeTags(msg.Config)
	w.db = notmuch.NewDB(pathToDB, excludedTags, w.w.Logger)
	return nil
}

func (w *worker) handleConnect(msg *types.Connect) error {
	err := w.db.Connect()
	if err != nil {
		return err
	}
	w.done(msg)
	w.emitLabelList()
	go func() {
		defer logging.PanicHandler()

		for {
			w.nmEvents <- &updateDirCounts{}
			time.Sleep(backgroundRefreshDelay)
		}
	}()
	return nil
}

func (w *worker) handleListDirectories(msg *types.ListDirectories) error {
	for _, name := range w.queryMapOrder {
		w.w.PostMessage(&types.Directory{
			Message: types.RespondTo(msg),
			Dir: &models.Directory{
				Name:       name,
				Attributes: []string{},
			},
		}, nil)
	}
	w.done(msg)
	return nil
}

func (w *worker) gatherDirectoryInfo(name string, query string) (
	*types.DirectoryInfo, error) {
	count, err := w.db.QueryCountMessages(query)
	if err != nil {
		return nil, err
	}
	info := &types.DirectoryInfo{
		Info: &models.DirectoryInfo{
			Name:     name,
			Flags:    []string{},
			ReadOnly: false,
			// total messages
			Exists: count.Exists,
			// new messages since mailbox was last opened
			Recent: 0,
			// total unread
			Unseen:         count.Unread,
			AccurateCounts: true,
		},
	}
	return info, nil
}

func (w *worker) emitDirectoryInfo(name string) error {
	query := w.queryFromName(name)
	info, err := w.gatherDirectoryInfo(name, query)
	if err != nil {
		return err
	}
	w.w.PostMessage(info, nil)
	return nil
}

//queryFromName either returns the friendly ID if aliased or the name itself
//assuming it to be the query
func (w *worker) queryFromName(name string) string {
	// try the friendly name first, if that fails assume it's a query
	q, ok := w.nameQueryMap[name]
	if !ok {
		return name
	}
	return q
}

func (w *worker) handleOpenDirectory(msg *types.OpenDirectory) error {
	w.w.Logger.Printf("opening %s", msg.Directory)
	// try the friendly name first, if that fails assume it's a query
	w.query = w.queryFromName(msg.Directory)
	w.currentQueryName = msg.Directory
	info, err := w.gatherDirectoryInfo(msg.Directory, w.query)
	if err != nil {
		return err
	}
	info.Message = types.RespondTo(msg)
	w.w.PostMessage(info, nil)
	w.done(msg)
	return nil
}

func (w *worker) handleFetchDirectoryContents(
	msg *types.FetchDirectoryContents) error {
	w.currentSortCriteria = msg.SortCriteria
	err := w.emitDirectoryContents(msg)
	if err != nil {
		return err
	}
	w.done(msg)
	return nil
}

func (w *worker) handleFetchDirectoryThreaded(
	msg *types.FetchDirectoryThreaded) error {
	// w.currentSortCriteria = msg.SortCriteria
	err := w.emitDirectoryThreaded(msg)
	if err != nil {
		return err
	}
	w.done(msg)
	return nil
}

func (w *worker) handleFetchMessageHeaders(
	msg *types.FetchMessageHeaders) error {
	for _, uid := range msg.Uids {
		m, err := w.msgFromUid(uid)
		if err != nil {
			w.w.Logger.Printf("could not get message: %v", err)
			w.err(msg, err)
			continue
		}
		err = w.emitMessageInfo(m, msg)
		if err != nil {
			w.w.Logger.Printf(err.Error())
			w.err(msg, err)
			continue
		}
	}
	w.done(msg)
	return nil
}

func (w *worker) uidsFromQuery(query string) ([]uint32, error) {
	msgIDs, err := w.db.MsgIDsFromQuery(query)
	if err != nil {
		return nil, err
	}
	var uids []uint32
	for _, id := range msgIDs {
		uid := w.db.UidFromKey(id)
		uids = append(uids, uid)

	}
	return uids, nil
}

func (w *worker) msgFromUid(uid uint32) (*Message, error) {
	key, ok := w.db.KeyFromUid(uid)
	if !ok {
		return nil, fmt.Errorf("Invalid uid: %v", uid)
	}
	msg := &Message{
		key: key,
		uid: uid,
		db:  w.db,
	}
	return msg, nil
}

func (w *worker) handleFetchMessageBodyPart(
	msg *types.FetchMessageBodyPart) error {

	m, err := w.msgFromUid(msg.Uid)
	if err != nil {
		w.w.Logger.Printf("could not get message %d: %v", msg.Uid, err)
		return err
	}
	r, err := m.NewBodyPartReader(msg.Part)
	if err != nil {
		w.w.Logger.Printf(
			"could not get body part reader for message=%d, parts=%#v: %v",
			msg.Uid, msg.Part, err)
		return err
	}
	w.w.PostMessage(&types.MessageBodyPart{
		Message: types.RespondTo(msg),
		Part: &models.MessageBodyPart{
			Reader: r,
			Uid:    msg.Uid,
		},
	}, nil)

	w.done(msg)
	return nil
}

func (w *worker) handleFetchFullMessages(msg *types.FetchFullMessages) error {
	for _, uid := range msg.Uids {
		m, err := w.msgFromUid(uid)
		if err != nil {
			w.w.Logger.Printf("could not get message %d: %v", uid, err)
			return err
		}
		r, err := m.NewReader()
		if err != nil {
			w.w.Logger.Printf("could not get message reader: %v", err)
			return err
		}
		defer r.Close()
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		w.w.PostMessage(&types.FullMessage{
			Message: types.RespondTo(msg),
			Content: &models.FullMessage{
				Uid:    uid,
				Reader: bytes.NewReader(b),
			},
		}, nil)
	}
	w.done(msg)
	return nil
}

func (w *worker) handleAnsweredMessages(msg *types.AnsweredMessages) error {
	for _, uid := range msg.Uids {
		m, err := w.msgFromUid(uid)
		if err != nil {
			w.w.Logger.Printf("could not get message: %v", err)
			w.err(msg, err)
			continue
		}
		if err := m.MarkAnswered(msg.Answered); err != nil {
			w.w.Logger.Printf("could not mark message as answered: %v", err)
			w.err(msg, err)
			continue
		}
		err = w.emitMessageInfo(m, msg)
		if err != nil {
			w.w.Logger.Printf(err.Error())
			w.err(msg, err)
			continue
		}
	}
	if err := w.emitDirectoryInfo(w.currentQueryName); err != nil {
		w.w.Logger.Printf(err.Error())
	}
	w.done(msg)
	return nil
}

func (w *worker) handleFlagMessages(msg *types.FlagMessages) error {
	for _, uid := range msg.Uids {
		m, err := w.msgFromUid(uid)
		if err != nil {
			w.w.Logger.Printf("could not get message: %v", err)
			w.err(msg, err)
			continue
		}
		if err := m.SetFlag(msg.Flag, msg.Enable); err != nil {
			w.w.Logger.Printf("could not set flag %v as %v for message: %v", msg.Flag, msg.Enable, err)
			w.err(msg, err)
			continue
		}
		err = w.emitMessageInfo(m, msg)
		if err != nil {
			w.w.Logger.Printf(err.Error())
			w.err(msg, err)
			continue
		}
	}
	if err := w.emitDirectoryInfo(w.currentQueryName); err != nil {
		w.w.Logger.Printf(err.Error())
	}
	w.done(msg)
	return nil
}

func (w *worker) handleSearchDirectory(msg *types.SearchDirectory) error {
	// the first item is the command (search / filter)
	s := strings.Join(msg.Argv[1:], " ")
	// we only want to search in the current query, so merge the two together
	search := w.query
	if s != "" {
		search = fmt.Sprintf("(%v) and (%v)", w.query, s)
	}
	uids, err := w.uidsFromQuery(search)
	if err != nil {
		return err
	}
	w.w.PostMessage(&types.SearchResults{
		Message: types.RespondTo(msg),
		Uids:    uids,
	}, nil)
	return nil
}

func (w *worker) handleModifyLabels(msg *types.ModifyLabels) error {
	for _, uid := range msg.Uids {
		m, err := w.msgFromUid(uid)
		if err != nil {
			return fmt.Errorf("could not get message from uid %v: %v", uid, err)
		}
		err = m.ModifyTags(msg.Add, msg.Remove)
		if err != nil {
			return fmt.Errorf("could not modify message tags: %v", err)
		}
		err = w.emitMessageInfo(m, msg)
		if err != nil {
			return err
		}
	}
	// tags changed, most probably some messages shifted to other folders
	// so we need to re-enumerate the query content
	err := w.emitDirectoryContents(msg)
	if err != nil {
		return err
	}
	// and update the list of possible tags
	w.emitLabelList()
	if err = w.emitDirectoryInfo(w.currentQueryName); err != nil {
		w.w.Logger.Printf(err.Error())
	}
	w.done(msg)
	return nil
}

func (w *worker) loadQueryMap(acctConfig *config.AccountConfig) error {
	raw, ok := acctConfig.Params["query-map"]
	if !ok {
		// nothing to do
		return nil
	}
	file, err := homedir.Expand(raw)
	if err != nil {
		return err
	}
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	w.nameQueryMap = make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || line[0] == '#' {
			continue
		}

		split := strings.SplitN(line, "=", 2)
		if len(split) != 2 {
			return fmt.Errorf("%v: invalid line %q, want name=query", file, line)
		}
		w.nameQueryMap[split[0]] = split[1]
		w.queryMapOrder = append(w.queryMapOrder, split[0])
	}
	return nil
}

func (w *worker) loadExcludeTags(
	acctConfig *config.AccountConfig) []string {
	raw, ok := acctConfig.Params["exclude-tags"]
	if !ok {
		// nothing to do
		return nil
	}
	excludedTags := strings.Split(raw, ",")
	for idx, tag := range excludedTags {
		excludedTags[idx] = strings.Trim(tag, " ")
	}
	return excludedTags
}

func (w *worker) emitDirectoryContents(parent types.WorkerMessage) error {
	uids, err := w.uidsFromQuery(w.query)
	if err != nil {
		return fmt.Errorf("could not fetch uids: %v", err)
	}
	sortedUids, err := w.sort(uids, w.currentSortCriteria)
	if err != nil {
		w.w.Logger.Printf("error sorting directory: %v", err)
		return err
	}
	w.w.PostMessage(&types.DirectoryContents{
		Message: types.RespondTo(parent),
		Uids:    sortedUids,
	}, nil)
	return nil
}

func (w *worker) emitDirectoryThreaded(parent types.WorkerMessage) error {
	threads, err := w.db.ThreadsFromQuery(w.query)
	if err != nil {
		return err
	}
	w.w.PostMessage(&types.DirectoryThreaded{
		Message: types.RespondTo(parent),
		Threads: threads,
	}, nil)
	return nil
}

func (w *worker) emitMessageInfo(m *Message,
	parent types.WorkerMessage) error {
	info, err := m.MessageInfo()
	if err != nil {
		return fmt.Errorf("could not get MessageInfo: %v", err)
	}
	w.w.PostMessage(&types.MessageInfo{
		Message: types.RespondTo(parent),
		Info:    info,
	}, nil)
	return nil
}

func (w *worker) emitLabelList() {
	tags, err := w.db.ListTags()
	if err != nil {
		w.w.Logger.Printf("could not load tags: %v", err)
		return
	}
	w.w.PostMessage(&types.LabelList{Labels: tags}, nil)
}

func (w *worker) sort(uids []uint32,
	criteria []*types.SortCriterion) ([]uint32, error) {
	if len(criteria) == 0 {
		return uids, nil
	}
	var msgInfos []*models.MessageInfo
	for _, uid := range uids {
		m, err := w.msgFromUid(uid)
		if err != nil {
			w.w.Logger.Printf("could not get message: %v", err)
			continue
		}
		info, err := m.MessageInfo()
		if err != nil {
			w.w.Logger.Printf("could not get message info: %v", err)
			continue
		}
		msgInfos = append(msgInfos, info)
	}
	sortedUids, err := lib.Sort(msgInfos, criteria)
	if err != nil {
		w.w.Logger.Printf("could not sort the messages: %v", err)
		return nil, err
	}
	return sortedUids, nil
}

func (w *worker) handleCheckMail(msg *types.CheckMail) {
	ctx, cancel := context.WithTimeout(context.Background(), msg.Timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", msg.Command)
	ch := make(chan error)
	go func() {
		err := cmd.Run()
		ch <- err
	}()
	select {
	case <-ctx.Done():
		w.err(msg, fmt.Errorf("checkmail: timed out"))
	case err := <-ch:
		if err != nil {
			w.err(msg, fmt.Errorf("checkmail: error running command: %v", err))
		} else {
			w.done(msg)
		}
	}
}
