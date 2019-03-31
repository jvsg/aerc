	"fmt"
	"github.com/emersion/go-imap"
	"git.sr.ht/~sircmpwn/aerc2/lib"
	"git.sr.ht/~sircmpwn/aerc2/worker/types"
	cmd    *exec.Cmd
	source io.Reader
	sink   io.WriteCloser
	grid   *ui.Grid
	term   *Terminal
}

func formatAddresses(addrs []*imap.Address) string {
	val := bytes.Buffer{}
	for i, addr := range addrs {
		if addr.PersonalName != "" {
			val.WriteString(fmt.Sprintf("%s <%s@%s>",
				addr.PersonalName, addr.MailboxName, addr.HostName))
		} else {
			val.WriteString(fmt.Sprintf("%s@%s",
				addr.MailboxName, addr.HostName))
		}
		if i != len(addrs)-1 {
			val.WriteString(", ")
		}
	}
	return val.String()
}

func NewMessageViewer(store *lib.MessageStore,
	msg *types.MessageInfo) *MessageViewer {

		{ui.SIZE_EXACT, 3}, // TODO: Based on number of header rows
	// TODO: let user specify additional headers to show by default
			Value: formatAddresses(msg.Envelope.From),
			Value: formatAddresses(msg.Envelope.To),
			Name:  "Subject",
			Value: msg.Envelope.Subject,
	headers.AddChild(ui.NewFill(' ')).At(2, 0).Span(1, 2)
	cmd := exec.Command("less")
	// TODO: configure multipart view. I left a spot for it in the grid
	body.AddChild(term).At(0, 0).Span(1, 2)

	grid.AddChild(headers).At(0, 0)
	grid.AddChild(body).At(1, 0)

	viewer := &MessageViewer{
		cmd:  cmd,
		sink: pipe,
		grid: grid,
		term: term,
	}

	store.FetchBodyPart(msg.Uid, 0, func(reader io.Reader) {
		viewer.source = reader
		viewer.attemptCopy()
	})

		viewer.attemptCopy()
	return viewer
}
func (mv *MessageViewer) attemptCopy() {
	if mv.source != nil && mv.cmd.Process != nil {
		go func() {
			io.Copy(mv.sink, mv.source)
			mv.sink.Close()
		}()
	}
	name := hv.Name
	size := runewidth.StringWidth(name)
	lim := ctx.Width() - size - 1
	value := runewidth.Truncate(" "+hv.Value, lim, "…")
	var (
		hstyle tcell.Style
		vstyle tcell.Style
	)
	// TODO: Make this more robust and less dumb
		vstyle = tcell.StyleDefault.Foreground(tcell.ColorGreen)
		hstyle = tcell.StyleDefault.Bold(true)
		vstyle = tcell.StyleDefault
		hstyle = tcell.StyleDefault.Bold(true)
	ctx.Fill(0, 0, ctx.Width(), ctx.Height(), ' ', vstyle)
	ctx.Printf(0, 0, hstyle, name)
	ctx.Printf(size, 0, vstyle, value)