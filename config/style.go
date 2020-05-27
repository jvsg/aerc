package config

import (
	"errors"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/go-ini/ini"
	"github.com/mitchellh/go-homedir"
)

type StyleObject int32

const (
	STYLE_DEFAULT StyleObject = iota
	STYLE_ERROR
	STYLE_WARNING
	STYLE_SUCCESS

	STYLE_TITLE
	STYLE_HEADER

	STYLE_STATUSLINE_DEFAULT
	STYLE_STATUSLINE_ERROR
	STYLE_STATUSLINE_SUCCESS

	STYLE_MSGLIST_DEFAULT
	STYLE_MSGLIST_UNREAD
	STYLE_MSGLIST_READ
	STYLE_MSGLIST_DELETED
	STYLE_MSGLIST_MARKED
	STYLE_MSGLIST_FLAGGED

	STYLE_DIRLIST_DEFAULT

	STYLE_COMPLETION_DEFAULT
	STYLE_COMPLETION_GUTTER
	STYLE_COMPLETION_PILL

	STYLE_TAB
	STYLE_STACK
	STYLE_SPINNER
	STYLE_BORDER

	STYLE_SELECTER_DEFAULT
	STYLE_SELECTER_FOCUSED
	STYLE_SELECTER_CHOOSER
)

var StyleNames = map[string]StyleObject{
	"default": STYLE_DEFAULT,
	"error":   STYLE_ERROR,
	"warning": STYLE_WARNING,
	"success": STYLE_SUCCESS,

	"title":  STYLE_TITLE,
	"header": STYLE_HEADER,

	"statusline_default": STYLE_STATUSLINE_DEFAULT,
	"statusline_error":   STYLE_STATUSLINE_ERROR,
	"statusline_success": STYLE_STATUSLINE_SUCCESS,

	"msglist_default": STYLE_MSGLIST_DEFAULT,
	"msglist_unread":  STYLE_MSGLIST_UNREAD,
	"msglist_read":    STYLE_MSGLIST_READ,
	"msglist_deleted": STYLE_MSGLIST_DELETED,
	"msglist_marked":  STYLE_MSGLIST_MARKED,
	"msglist_flagged": STYLE_MSGLIST_FLAGGED,

	"dirlist_default": STYLE_DIRLIST_DEFAULT,

	"completion_default": STYLE_COMPLETION_DEFAULT,
	"completion_gutter":  STYLE_COMPLETION_GUTTER,
	"completion_pill":    STYLE_COMPLETION_PILL,

	"tab":     STYLE_TAB,
	"stack":   STYLE_STACK,
	"spinner": STYLE_SPINNER,
	"border":  STYLE_BORDER,

	"selecter_default": STYLE_SELECTER_DEFAULT,
	"selecter_focused": STYLE_SELECTER_FOCUSED,
	"selecter_chooser": STYLE_SELECTER_CHOOSER,
}

type Style struct {
	Fg        tcell.Color
	Bg        tcell.Color
	Bold      bool
	Blink     bool
	Underline bool
	Reverse   bool
}

func (s Style) Get() tcell.Style {
	return tcell.StyleDefault.
		Foreground(s.Fg).
		Background(s.Bg).
		Bold(s.Bold).
		Blink(s.Blink).
		Underline(s.Blink).
		Reverse(s.Reverse)
}

func (s *Style) Normal() {
	s.Bold = false
	s.Blink = false
	s.Underline = false
	s.Reverse = false
}

func (s *Style) Default() *Style {
	s.Fg = tcell.ColorDefault
	s.Bg = tcell.ColorDefault
	return s
}

func (s *Style) Reset() *Style {
	s.Default()
	s.Normal()
	return s
}

func boolSwitch(val string, cur_val bool) (bool, error) {
	switch val {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "toggle":
		return !cur_val, nil
	default:
		return cur_val, errors.New(
			"Bool Switch attribute must be true, false, or toggle")
	}
}

func (s *Style) Set(attr, val string) error {
	switch attr {
	case "fg":
		s.Fg = tcell.GetColor(val)
	case "bg":
		s.Bg = tcell.GetColor(val)
	case "bold":
		if state, err := boolSwitch(val, s.Bold); err != nil {
			return err
		} else {
			s.Bold = state
		}
	case "blink":
		if state, err := boolSwitch(val, s.Blink); err != nil {
			return err
		} else {
			s.Blink = state
		}
	case "underline":
		if state, err := boolSwitch(val, s.Underline); err != nil {
			return err
		} else {
			s.Underline = state
		}
	case "reverse":
		if state, err := boolSwitch(val, s.Reverse); err != nil {
			return err
		} else {
			s.Reverse = state
		}
	case "default":
		s.Default()
	case "normal":
		s.Normal()
	default:
		return errors.New("Unknown style attribute: " + attr)
	}

	return nil
}

type StyleSet struct {
	objects  map[StyleObject]*Style
	selected map[StyleObject]*Style
}

func NewStyleSet() StyleSet {
	ss := StyleSet{
		objects:  make(map[StyleObject]*Style),
		selected: make(map[StyleObject]*Style),
	}
	for _, so := range StyleNames {
		ss.objects[so] = new(Style)
		ss.selected[so] = new(Style)
	}

	return ss
}

func (ss StyleSet) reset() {
	for _, so := range StyleNames {
		ss.objects[so].Reset()
		ss.selected[so].Reset()
	}
}

func (ss StyleSet) Get(so StyleObject) tcell.Style {
	return ss.objects[so].Get()
}

func (ss StyleSet) Selected(so StyleObject) tcell.Style {
	return ss.selected[so].Get()
}

func findStyleSet(stylesetName string, stylesetsDir []string) (string, error) {
	for _, dir := range stylesetsDir {
		stylesetPath, err := homedir.Expand(path.Join(dir, stylesetName))
		if err != nil {
			return "", err
		}

		if _, err := os.Stat(stylesetPath); os.IsNotExist(err) {
			continue
		}

		return stylesetPath, nil
	}

	return "", errors.New("Can't find styleset - " + stylesetName)
}
func (ss *StyleSet) ParseStyleSet(stylesetName string, stylesetDirs []string) error {
	filepath, err := findStyleSet(stylesetName, stylesetDirs)
	if err != nil {
		return err
	}

	file, err := ini.Load(filepath)
	if err != nil {
		return err
	}

	ss.reset()

	defaultSection, err := file.GetSection(ini.DefaultSection)
	if err != nil {
		return err
	}

	selectedKeys := []string{}

	for _, key := range defaultSection.KeyStrings() {
		tokens := strings.Split(key, ".")
		var styleName, attr string
		switch len(tokens) {
		case 2:
			styleName, attr = tokens[0], tokens[1]
		case 3:
			if tokens[1] != "selected" {
				return errors.New("Unknown modifier: " + tokens[1])
			}
			selectedKeys = append(selectedKeys, key)
			continue
		default:
			return errors.New("Style parsing error: " + key)
		}
		val := defaultSection.KeysHash()[key]

		if strings.ContainsAny(styleName, "*?") {
			regex := fnmatchToRegex(styleName)
			for sn, so := range StyleNames {
				matched, err := regexp.MatchString(regex, sn)
				if err != nil {
					return err
				}

				if !matched {
					continue
				}

				if err := ss.objects[so].Set(attr, val); err != nil {
					return err
				}
				if err := ss.selected[so].Set(attr, val); err != nil {
					return err
				}
			}
		} else {
			so, ok := StyleNames[styleName]
			if !ok {
				return errors.New("Unknown style object: " + styleName)
			}
			if err := ss.objects[so].Set(attr, val); err != nil {
				return err
			}
			if err := ss.selected[so].Set(attr, val); err != nil {
				return err
			}
		}
	}

	for _, key := range selectedKeys {
		tokens := strings.Split(key, ".")
		styleName, modifier, attr := tokens[0], tokens[1], tokens[2]
		if modifier != "selected" {
			return errors.New("Unknown modifier: " + modifier)
		}

		val := defaultSection.KeysHash()[key]

		if strings.ContainsAny(styleName, "*?") {
			regex := fnmatchToRegex(styleName)
			for sn, so := range StyleNames {
				matched, err := regexp.MatchString(regex, sn)
				if err != nil {
					return err
				}

				if !matched {
					continue
				}

				if err := ss.selected[so].Set(attr, val); err != nil {
					return err
				}
			}
		} else {
			so, ok := StyleNames[styleName]
			if !ok {
				return errors.New("Unknown style object: " + styleName)
			}
			if err := ss.selected[so].Set(attr, val); err != nil {
				return err
			}
		}
	}

	for _, key := range defaultSection.KeyStrings() {
		tokens := strings.Split(key, ".")
		styleName, attr := tokens[0], tokens[1]
		val := defaultSection.KeysHash()[key]

		if styleName != "selected" {
			continue
		}

		for _, so := range StyleNames {
			if err := ss.selected[so].Set(attr, val); err != nil {
				return err
			}
		}
	}

	return nil
}

func fnmatchToRegex(pattern string) string {
	n := len(pattern)
	var regex strings.Builder

	for i := 0; i < n; i++ {
		switch pattern[i] {
		case '*':
			regex.WriteString(".*")
		case '?':
			regex.WriteByte('.')
		default:
			regex.WriteByte(pattern[i])
		}
	}

	return regex.String()
}
