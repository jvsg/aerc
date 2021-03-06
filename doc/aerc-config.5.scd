aerc-config(5)

# NAME

aerc-config - configuration file formats for *aerc*(1)

# CONFIGURATION

There are three aerc config files: *aerc.conf*, *binds.conf*, and
*accounts.conf*. The last one must be kept secret, as it may include your
account credentials. We look for these files in your XDG config home plus
"aerc", which defaults to ~/.config/aerc.

Examples of these config files are typically included with your installation of
aerc and are usually installed in /usr/share/aerc.

Each file uses the _ini_ format, and consists of sections with keys and values.
A line beginning with # is considered a comment and ignored, as are empty lines.
New sections begin with [section-name] on a single line, and keys and values are
separated with "=".

# AERC.CONF

This file is used for configuring the general appearance and behavior of aerc.

## GENERAL OPTIONS

These options are configured in the *[general]* section of aerc.conf.

*default-save-path*
	Used as a default path for save operations if no other path is specified.

*pgp-provider*
	If set to "gpg", aerc will use system gpg binary and keystore for all
	crypto operations. Otherwise, the internal openpgp implemenation will be
	used.

	Default: internal

*unsafe-accounts-conf*
	By default, the file permissions of accounts.conf must be restrictive
	and only allow reading by the file owner (_0600_). Set this option to
	*true* to ignore this permission check. Use this with care as it may
	expose your credentials.

	Default: false

## UI OPTIONS

These options are configured in the *[ui]* section of aerc.conf.

*index-format*
	Describes the format for each row in a mailbox view. This field is
	compatible with mutt's printf-like syntax.

	Default: %D %-17.17n %s

[- *Format specifier*
:[ *Description*
|  %%
:  literal %
|  %a
:  sender address
|  %A
:  reply-to address, or sender address if none
|  %C
:  message number
|  %d
:  formatted message timestamp
|  %D
:  formatted message timestamp converted to local timezone
|  %f
:  sender name and address
|  %F
:  author name, or recipient name if the message is from you.
   The address is shown if no name part.
|  %g
:  message labels (for example notmuch tags)
|  %i
:  message id
|  %n
:  sender name, or sender address if none
|  %r
:  comma-separated list of formatted recipient names and addresses
|  %R
:  comma-separated list of formatted CC names and addresses
|  %s
:  subject
|  %t
:  the (first) address the new email was sent to
|  %T
:  the account name which received the email
|  %u
:  sender mailbox name (e.g. "smith" in "smith@example.net")
|  %v
:  sender first name (e.g. "Alex" in "Alex Smith <smith@example.net>")
|  %Z
:  flags (O=old, N=new, r=answered, D=deleted, !=flagged, \*=marked)

*timestamp-format*
	See time.Time#Format at https://godoc.org/time#Time.Format

	Default: "2006-01-02 03:04 PM" (ISO 8601 + 12 hour time)

*this-day-time-format*
	Index-only time format for messages that were received/sent today.
	If this is not specified, *timestamp-format* is used instead.

	Default: "03:04 PM" (12 hour time)

*this-week-time-format*
	Index-only time format for messages that were received/sent within the
	last 7 days. If this is not specified, *timestamp-format* is used
	instead.

	Default: "Monday 03:04 PM" (Week day + 12 hour time)

*this-year-time-format*
	Index-only time format for messages that were received/sent this year.
	If this is not specified, *timestamp-format* is used instead.

	Default: "January 02" (Month + month day)

*sidebar-width*
	Width of the sidebar, including the border. Set to zero to disable the
	sidebar.

	Default: 20

*empty-message*
	Message to display when viewing an empty folder.

	Default: (no messages)

*empty-dirlist*
	Message to display when no folders exists or are all filtered.

	Default: (no folders)

*mouse-enabled*
	Enable mouse events in the ui, e.g. clicking and scrolling with the mousewheel

	Default: false

*new-message-bell*
	Ring the bell when a new message is received.

	Default: true

*pinned-tab-marker*
	Marker to show before a pinned tab's name.

	Default: `

*spinner*
	Animation shown while loading, split by spinner-delimiter (below)

	Examples:
	- spinner = "\-\_-,\_-\_"
	- spinner = '. , .'
	- spinner = "\,|,/,-"

	Default: "[..]    , [..]   ,  [..]  ,   [..] ,    [..],   [..] ,  [..]  , [..]   "

*spinner-delimiter*
	Spinner delimiter to split string into an animation

	Default: ","

*sort*
	List of space-separated criteria to sort the messages by, see *sort*
	command in *aerc*(1) for reference. Prefixing a criterion with "-r "
	reverses that criterion.

	Example: "from -r date"

	Default: ""

*dirlist-format*
	Describes the format string to use for the directory list

	Default: %n %>r

[- *Format specifier*
:[ *Description*
|  %%
:  literal %
|  %n
:  directory name
|  %N
:  compacted directory name
|  %r
:  recent/unseen/total message count
|  %>X
:  make format specifier 'X' be right justified

*dirlist-delay*
	Delay after which the messages are actually listed when entering
	a directory. This avoids loading messages when skipping over folders
	and makes the UI more responsive. If you do not want that, set it to
	0s.

	Default: 200ms

*dirlist-tree*
	Display the directory list as a foldable tree that allows to collapse
	and expand the folders.

	Default: false

*next-message-on-delete*
	Moves to next message when the current message is deleted

	Default: true

*completion-popovers*
	Shows potential auto-completions for text inputs in popovers.

	Default: true

*completion-delay*
	How long to wait after the last input before auto-completion is triggered.

	Default: 250ms

*border-char-vertical*
*border-char-horizontal*
	Set stylable characters (via the 'border' element) for vertical and
	horizontal borders.

	Default: spaces

*stylesets-dirs*
	The directories where the stylesets are stored. The config takes
	a colon-separated list of dirs. If this is unset or if a styleset cannot
	be found, the following paths will be used as a fallback in that order:

	```
	${XDG_CONFIG_HOME:-~/.config}/aerc/stylesets
	${XDG_DATA_HOME:-~/.local/share}/aerc/stylesets
	/usr/local/share/aerc/stylesets
	/usr/share/aerc/stylesets
	```

	Default: ""

*styleset-name*
	The name of the styleset to be used to style the ui elements. The
	stylesets are stored in the 'stylesets' directory in the config
	directory.

	Default: default

	Have a look at *aerc-stylesets*(7) as to how a styleset looks like.

*fuzzy-complete*
	When typing a command or option, the popover will now show not only the
	items /starting/ with the string input by the user, but it will also show
	instances of items /containing/ the string, starting at any position and
	need not be consecutive characters in the command or option.

*threading-enabled*
	Enable a threaded viewing of messages, works with IMAP (when there's
	server support) and NotMuch backends.

	This option should only be set to true for specific accounts
	accordingly. See *Contextual UI Configuration* below.

	If this is not supported by the server, you can enable it on the fly by
	issuing the _:toggle-threads_ command. The threading will be done client
	side instead of on the server.

	Default: false


## Contextual UI Configuration

The UI configuration can be specialized for accounts, specific mail
directories and message subjects. The specializations are added using
contextual config sections based on the context.

The contextual UI configuration is merged to the base UiConfig in the
following order:
*Base UIConfig > Account Context > Folder Context > Subject Context.*

*[ui:account=<AccountName>]*
	Adds account specific configuration with the account name.

*[ui:folder=<FolderName>]*
	Add folder specific configuration with the folder name.

*[ui:folder~<Regex>]*
	Add folder specific configuration for folders whose names match the regular
	expression.

*[ui:subject~<Regex>]*
	Add specialized ui configuration for messages that match a given regular
	expression.

Example:
```
[ui:account=Work]
sidebar-width=...

[ui:folder=Sent]
index-format=...

[ui:folder~Archive/\d+/.*]
index-format=...

[ui:subject~^\[PATCH]
index-format=...
```

## STATUSLINE

These options are configured in the *[statusline]* section of aerc.conf.

*render-format*
	Describes the format string for the statusline format.

	For a minimal statusline that only shows the current account and
	the connection information, use [%a] %c.

	To completely mute the statusline (except for push notficiations), use
	%m only.

	Default: [%a] %S %>%T

[- *Format specifier*
:[ *Description*
|  %%
:  literal %
|  %a
:  active account name
|  %d
:  active directory name
|  %c
:  connection state
|  %m
:  mute statusline and show only push notifications
|  %S
:  general status information (e.g. connection state, filter, search)
|  %T
:  general on/off information (e.g. passthrough, threading, sorting)
|  %>
:  does not print anything but all format specifier that follow will be right justified. 

*separator*
	Specifies the separator between grouped statusline elements (e.g. for
	the %S and %T specifiers in *render-format*).

	Default: " | "

*display-mode*
	Defines the mode for displaying the status elements.
	Options: text, icon

	Default: text


## VIEWER

These options are configured in the *[viewer]* section of aerc.conf.

*pager*
	Specifies the pager to use when displaying emails. Note that some filters
	may add ANSI escape sequences to add color to rendered emails, so you may
	want to use a pager which supports ANSI.

	Default: less -R

*alternatives*
	If an email offers several versions (multipart), you can configure which
	mimetype to prefer. For example, this can be used to prefer plaintext over
	HTML emails.

	Default: text/plain,text/html

*header-layout*
	Defines the default headers to display when viewing a message. To display
	multiple headers in the same row, separate them with a pipe, e.g. "From|To".
	Rows will be hidden if none of their specified headers are present in the
	message.

	Authentication information from the Authentication-Results header can be
	displayed by adding DKIM, SPF or DMARC. To show more information
	than just the authentication result, append a plus sign (+) to the header name
	(e.g. DKIM+).

	Default: From|To,Cc|Bcc,Date,Subject

*show-headers*
	Default setting to determine whether to show full headers or only parsed
	ones in message viewer.

	Default: false

*always-show-mime*
	Whether to always show the mimetype of an email, even when it is just a single part.

	Default: false

*parse-http-links*
	Parses and extracts http links when viewing a message. Links can then be
	accessed with the open-link command.

	Default: true

## COMPOSE

These options are configured in the *[compose]* section of aerc.conf.

*editor*
	Specifies the command to run the editor with. It will be shown in an
	embedded terminal, though it may also launch a graphical window if the
	environment supports it. Defaults to *$EDITOR*, or *vi*(1).

*header-layout*
	Defines the default headers to display when composing a message. To display
	multiple headers in the same row, separate them with a pipe, e.g. "To|From".

	Default: To|From,Subject

*address-book-cmd*
	Specifies the command to be used to tab-complete email addresses. Any
	occurrence of "%s" in the address-book-cmd will be replaced with anything
	the user has typed after the last comma.

	The command must output the completions to standard output, one completion
	per line. Each line must be tab-delimited, with an email address occurring as
	the first field. Only the email address field is required. The second field,
	if present, will be treated as the contact name. Additional fields are
	ignored.

	Example:
		khard email --remove-first-line --parsable '%s'

	Default: none

*reply-to-self*
	Allow to include your own address. Otherwise, if set to false, do not mail
	yourself when replying (e.g., if replying to emails previously sent by
	yourself, address your replies to the original To: and Cc:).

	Default: true

## FILTERS

Filters allow you to pipe an email body through a shell command to render
certain emails differently, e.g. highlighting them with ANSI escape codes.
They are configured in the *[filters]* section of aerc.conf.

The first filter which matches the email's mimetype will be used, so order
them from most to least specific.

You can also match on non-mimetypes, by prefixing with the header to match
against (non-case-sensitive) and a comma, e.g. subject,text will match a
subject which contains "text". Use header,~regex to match against a regex.

aerc ships with some default filters installed in the share directory (usually
_/usr/share/aerc/filters_). Note that these may have additional dependencies
that aerc does not have alone.

## TRIGGERS

Triggers specify commands to execute when certain events occur.

They are configured in the *[triggers]* section of aerc.conf.

*new-email*
	Executed when a new email arrives in the selected folder.

	e.g. new-email=exec notify-send "New email from %n" "%s"

	Default: ""

	Format specifiers from *index-format* are expanded with respect to the new
	message.

## TEMPLATES

Templates are used to populate the body of an email. The compose, reply
and forward commands can be called with the -T flag with the name of the
template name.

aerc ships with some default templates installed in the share directory (usually
_/usr/share/aerc/templates_).

These options are configured in the *[templates]* section of aerc.conf.

*template-dirs*
	The directory where the templates are stored. The config takes
	a colon-separated list of dirs. If this is unset or if a template cannot
	be found, the following paths will be used as a fallback in that order:

	```
	${XDG_CONFIG_HOME:-~/.config}/aerc/templates
	${XDG_DATA_HOME:-~/.local/share}/aerc/templates
	/usr/local/share/aerc/templates
	/usr/share/aerc/templates
	```

	Default: ""

*new-message*
	The default template to be used for new messages.

	Default: "new_message"

*quoted-reply*
	The default template to be used for quoted replies.

	Default: "quoted_reply"

*forwards*
	The default template to be used for forward as body.

	Default: "forward_as_body"

# ACCOUNTS.CONF

This file is used for configuring each mail account used for aerc. Each section
is the name of an account you want to configure, and the keys & values in that
section specify details of that account's configuration. In addition to the
options documented here, specific transports for incoming and outgoing emails
may have additional configuration parameters, documented on their respective man
pages.

Note that many of these configuration options are written for you, such as
*source* and *outgoing*, when you run the account configuration wizard
(*:new-account*).

*archive*
	Specifies a folder to use as the destination of the *:archive* command.

	Default: Archive

*check-mail*
	Specifies an interval to check for new mail. Mail will be checked at
	startup, and every interval. IMAP accounts will check for mail in all
	unselected folders, and the selected folder will continue to receive PUSH
	mail notifications. Maildir/Notmuch folders must use *check-mail-cmd* in
	conjunction with this option. See *aerc-maildir* and *aerc-notmuch* for
	more information.

	Setting this option to 0 will disable check-mail

	Example: 5m

	Default: 0

*copy-to*
	Specifies a folder to copy sent mails to, usually "Sent".

	Default: none

*default*
	Specifies the default folder to open in the message list when aerc
	configures this account.

	Default: INBOX

*folders*
	Specifies the comma separated list of folders to display in the sidebar.
	Names prefixed with ~ are interpreted as regular expressions.

	Default: all folders

*folders-exclude*
	Specifies the comma separated list of folders to exclude from the sidebar.
	Names prefixed with ~ are interpreted as regular expressions.
	Note that this overrides anything from *folders*.

	Default: no folders

*enable-folders-sort*
	If true, folders are sorted, first by specified folders (see *folders-sort*),
	then alphabetically.

	Default: true

*folders-sort*
	Specifies a comma separated list of folders to be shown at the top of the
	list in the provided order. Remaining folders will be sorted alphabetically.

	Default: none

*from*
	The default value to use for the From header in new emails. This should be
	an RFC 5322-compatible string, such as "Your Name <you@example.org>".

	Default: none

*aliases*
	All aliases of the current account. These will be used to fill in the From:
	field. Make sure that your email server accepts this value, or for example
	use *aerc-sendmail*(5) in combination with msmtp and --read-envelope-from.

	Default: none

*outgoing*
	Specifies the transport for sending outgoing emails on this account.  It
	should be a connection string, and the specific meaning of each component
	varies depending on the protocol in use.  See each protocol's man page for
	more details:

	- *aerc-smtp*(5)

*outgoing-cred-cmd*
	Specifies an optional command that is run to get the outgoing account's
	password. See each protocol's man page for more details.

	Default: none

*pgp-auto-sign*
	If true, all outgoing emails from this account will be signed (if a signing
	key is available)

	Default: false

*pgp-key-id*
	Specify the key id to use when signing a message. Can be either short or
	long key id. If unset, aerc will look up the key by email

*pgp-opportunistic-encrypt*
	If true, any outgoing email from this account will be encrypted when all
	recipients (including "cc" and "bcc" field) have a public key available in
	the keyring

	Default: false

*postpone*
	Specifies the folder to save postponed messages to.

	Default: Drafts

*source*
	Specifies the source for reading incoming emails on this account. This key
	is required for all accounts. It should be a connection string, and the
	specific meaning of each component varies depending on the protocol in use.
	See each protocol's man page for more details:

	- *aerc-imap*(5)
	- *aerc-maildir*(5)
	- *aerc-notmuch*(5)

	Default: none

*source-cred-cmd*
	Specifies an optional command that is run to get the source account's
	password. See each protocol's man page for more details.

*signature-file*
	Specifies the file to read in order to obtain the signature to be added
	to emails sent from this account.

*signature-cmd*
	Specifies the command to execute in *sh* in order to obtain the
	signature to be added to emails sent from this account. If the command
	fails then *signature-file* is used instead.

*trusted-authres*
	Comma-separated list of trustworthy hostnames from which the
	Authentication Results header will be displayed. Entries can be regular
	expressions. If you want to trust any host (e.g. for debugging),
	use the wildcard \*.

# BINDS.CONF

This file is used for configuring keybindings used in the aerc interactive
client. You may configure different keybindings for different contexts by
writing them into different *[sections]* of the ini file. The available contexts
are:

*[messages]*
	keybindings for the message list

*[view]*
	keybindings for the message viewer

*[view::passthrough]*
	keybindings for the viewer, when in key passthrough mode
	(toggled with :toggle-key-passthrough)

*[compose]*
	keybindings for the message composer

*[compose::editor]*
	keybindings for the composer, when the editor is focused

*[compose::review]*
	keybindings for the composer, when reviewing the email before it's sent

*[terminal]*
	keybindings for terminal tabs

You may also configure account specific key bindings for each context:

*[context:account=<AccountName>]*
	keybindings for this context and account, where <AccountName> matches
	the account name you provided in *accounts.conf*.

Example:
```
[messages:account=Mailbox]
c = :cf path:mailbox/** and<space>

[compose::editor:account=Mailbox2]
...
```

You may also configure global keybindings by placing them at the beginning of
the file, before specifying any context-specific sections. For each *key=value*
option specified, the _key_ is the keystrokes pressed (in order) to invoke this
keybinding, and _value_ specifies keystrokes that aerc will simulate when the
keybinding is invoked. Generally this is used to execute commands, for example:

	rq = :reply -q<Enter>

Pressing r, then q, will simulate typing in ":reply -q<Enter>", and execute
:reply -q accordingly. It is also possible to invoke keybindings recursively in
a similar fashion. Additionally, the following special options are available in
each binding context:

*$noinherit*
	If set to "true", global keybindings will not be effective in this context.

	Default: false

*$ex*
	This can be set to a keystroke which will bring up the command input in this
	context.
	
	Default: <semicolon>

In addition to letters, special keys may be specified in <angle brackets>. The
following special keys are supported:

[[ *Name*
:- *Description*
|  space
:  " "
|  semicolon
:  ;
|  tab
:  
|  enter
:  
|  up
:  
|  c-up
:  Ctrl+Up
|  down
:  
|  c-down
:  Ctrl+Down
|  right
:  
|  c-right
:  Ctrl+Right
|  left
:  
|  c-left
:  Ctrl+Left
|  pgup
:  
|  c-pgup
:  Ctrl+PageUp
|  pgdn
:  
|  c-pgdn
:  Ctrl+PageUp
|  home
:  
|  end
:  
|  insert
:  
|  delete
:  
|  exit
:  
|  cancel
:  
|  print
:  
|  pause
:  
|  backtab
: 
|  c-space
:  Ctrl+Space
|  c-a
:  Ctrl+a
|  c-b
:  Ctrl+b
|  c-c
:  Ctrl+c
|  c-d
:  Ctrl+d
|  c-e
:  Ctrl+e
|  c-f
:  Ctrl+f
|  c-g
:  Ctrl+g
|  c-h
:  Ctrl+h
|  c-i
:  Ctrl+i
|  c-j
:  Ctrl+j
|  c-k
:  Ctrl+k
|  c-l
:  Ctrl+l
|  c-m
:  Ctrl+m
|  c-n
:  Ctrl+n
|  c-o
:  Ctrl+o
|  c-p
:  Ctrl+p
|  c-q
:  Ctrl+q
|  c-r
:  Ctrl+r
|  c-s
:  Ctrl+s
|  c-t
:  Ctrl+t
|  c-u
:  Ctrl+u
|  c-v
:  Ctrl+v
|  c-w
:  Ctrl+w
|  c-x
:  Ctrl+x
|  c-y
:  Ctrl+y
|  c-z
:  Ctrl+z
|  c-]
:  Ctrl+]
|  c-[
:  Ctrl+[
|  c-^
:  Ctrl+^
|  c-\_
:  Ctrl+\_

# SEE ALSO

*aerc*(1) *aerc-imap*(5) *aerc-smtp*(5) *aerc-maildir*(5) *aerc-sendmail*(5)
*aerc-notmuch*(5) *aerc-stylesets*(7)

# AUTHORS

Originally created by Drew DeVault <sir@cmpwn.com> and maintained by Robin
Jarry <robin@jarry.cc> who is assisted by other open source contributors. For
more information about aerc development, see https://sr.ht/~rjarry/aerc/.
