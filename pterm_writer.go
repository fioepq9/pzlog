package pzlog

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"io"

	"github.com/gookit/color"
	"github.com/pterm/pterm"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type field struct {
	Key string
	Val string
}

type PtermWriter struct {
	MaxWidth        int
	Out             io.Writer
	TimeStyle       *pterm.Style
	TraceLevelStyle *pterm.Style
	DebugLevelStyle *pterm.Style
	InfoLevelStyle  *pterm.Style
	WarnLevelStyle  *pterm.Style
	ErrorLevelStyle *pterm.Style
	FatalLevelStyle *pterm.Style
	PanicLevelStyle *pterm.Style
	NoLevelStyle    *pterm.Style
	KeyStyles       map[string]*pterm.Style
	ValFormatters   map[string]func(padding int, v interface{}) string
	Less            func(string, string) bool
}

func NewPtermWriter(options ...func(*PtermWriter)) zerolog.LevelWriter {
	pw := PtermWriter{
		MaxWidth:        pterm.GetTerminalWidth(),
		Out:             os.Stdout,
		TimeStyle:       pterm.NewStyle(pterm.FgGray),
		TraceLevelStyle: pterm.NewStyle(pterm.Bold, pterm.FgCyan),
		DebugLevelStyle: pterm.NewStyle(pterm.Bold, pterm.FgBlue),
		InfoLevelStyle:  pterm.NewStyle(pterm.Bold, pterm.FgGreen),
		WarnLevelStyle:  pterm.NewStyle(pterm.Bold, pterm.FgYellow),
		ErrorLevelStyle: pterm.NewStyle(pterm.Bold, pterm.FgRed),
		FatalLevelStyle: pterm.NewStyle(pterm.Bold, pterm.FgRed),
		PanicLevelStyle: pterm.NewStyle(pterm.Bold, pterm.FgRed),
		NoLevelStyle:    pterm.NewStyle(pterm.Bold, pterm.FgWhite),
		KeyStyles: map[string]*pterm.Style{
			zerolog.CallerFieldName:     pterm.NewStyle(pterm.FgGray),
			zerolog.ErrorFieldName:      pterm.NewStyle(pterm.FgRed),
			zerolog.ErrorStackFieldName: pterm.NewStyle(pterm.FgRed),
		},
		ValFormatters: map[string]func(padding int, v interface{}) string{
			zerolog.ErrorStackFieldName: TreeFormatter,
		},
		Less: func(s1, s2 string) bool {
			score := func(s string) string {
				if s == zerolog.CallerFieldName {
					return string([]byte{0})
				}
				if s == zerolog.ErrorFieldName {
					return string([]byte{math.MaxUint8, 0})
				}
				if s == zerolog.ErrorStackFieldName {
					return string([]byte{math.MaxUint8, 1})
				}
				return s
			}
			return score(s1) < score(s2)
		},
	}

	for _, opt := range options {
		opt(&pw)
	}

	return &pw
}

func (pw *PtermWriter) Write(p []byte) (n int, err error) {
	return pw.Out.Write(p)
}

func (pw *PtermWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	var evt map[string]interface{}
	d := json.NewDecoder(bytes.NewReader(p))
	d.UseNumber()
	err = d.Decode(&evt)
	if err != nil {
		return n, fmt.Errorf("cannot decode event: %s", err)
	}

	var buf bytes.Buffer

	// timestamp
	if timestamp, ok := evt[zerolog.TimestampFieldName]; ok {
		delete(evt, zerolog.TimestampFieldName)
		buf.WriteString(pw.TimeStyle.Sprint(timestamp) + " ")
	}

	// level
	delete(evt, zerolog.LevelFieldName)
	var lvlStyle *pterm.Style
	switch level {
	case zerolog.TraceLevel:
		lvlStyle = pw.TraceLevelStyle
	case zerolog.DebugLevel:
		lvlStyle = pw.DebugLevelStyle
	case zerolog.InfoLevel:
		lvlStyle = pw.InfoLevelStyle
	case zerolog.WarnLevel:
		lvlStyle = pw.WarnLevelStyle
	case zerolog.ErrorLevel:
		lvlStyle = pw.ErrorLevelStyle
	case zerolog.FatalLevel:
		lvlStyle = pw.FatalLevelStyle
	case zerolog.PanicLevel:
		lvlStyle = pw.PanicLevelStyle
	case zerolog.NoLevel:
		lvlStyle = pw.NoLevelStyle
	case zerolog.Disabled:
		lvlStyle = pw.NoLevelStyle
	default:
		lvlStyle = pw.NoLevelStyle
	}
	buf.WriteString(lvlStyle.Sprintf("%+5s", strings.ToUpper(level.String())))
	padding := len(color.ClearCode(buf.String())) - 3

	// message
	if message, ok := evt[zerolog.MessageFieldName]; ok {
		delete(evt, zerolog.MessageFieldName)
		terminalWidth := lo.Min([]int{pterm.GetTerminalWidth(), pw.MaxWidth})
		remainingWidth := terminalWidth - len(color.ClearCode(buf.String())) - 2
		lines := strings.Split(pterm.Sprint(message), "\n")
		lines = lo.Filter(lines, func(line string, _ int) bool {
			return len(line) != 0
		})
		sep := "\n" + strings.Repeat(" ", padding) + "│     "
		switch len(lines) {
		case 0:
			buf.WriteByte('\n')
		case 1:
			msg := "  " + lines[0]
			if len(msg) > remainingWidth {
				msg = pterm.DefaultParagraph.WithMaxWidth(remainingWidth).Sprint(msg)
				msg = "  " + strings.ReplaceAll(msg, "\n", sep)
				buf.WriteString(msg)
				buf.WriteString("\n" + strings.Repeat(" ", padding) + "└" + strings.Repeat("-", remainingWidth))
			} else {
				buf.WriteString(msg)
			}
		default:
			for i, msg := range lines {
				if len(msg) > remainingWidth {
					msg = pterm.DefaultParagraph.WithMaxWidth(remainingWidth).Sprint(msg)
					msg = "  " + strings.ReplaceAll(msg, "\n", sep)
				} else if i == 0 {
					msg = "  " + msg
				} else {
					msg = "  " + sep + msg
				}
				buf.WriteString(msg)
			}
			buf.WriteString("\n" + strings.Repeat(" ", padding) + "└" + strings.Repeat("-", remainingWidth))
		}
	}

	// fields
	fields := make([]field, 0)
	for k, v := range evt {
		fields = append(fields, field{
			Key: k,
			Val: pw.Ksprint(lvlStyle, k) + pw.Vsprint(padding, k, v),
		})
	}
	sort.Slice(fields, func(i, j int) bool {
		return pw.Less(fields[i].Key, fields[j].Key)
	})

	for i, field := range fields {
		var pipe string
		if i < len(fields)-1 {
			pipe = "├"
		} else {
			pipe = "└"
		}
		buf.WriteString("\n" + strings.Repeat(" ", padding) + pipe + " " + field.Val)
	}

	buf.WriteByte('\n')
	_, err = buf.WriteTo(pw.Out)
	return n, err
}

func (pw *PtermWriter) Ksprint(defaultStyle *pterm.Style, k string) string {
	if style, ok := pw.KeyStyles[k]; ok {
		return style.Sprint(k)
	}
	return defaultStyle.Sprint(k)
}

func (pw *PtermWriter) Vsprint(padding int, k string, v interface{}) string {
	if formatter, ok := pw.ValFormatters[k]; ok {
		return formatter(padding, v)
	}
	return pterm.Sprintf(": %v", v)
}

func TreeFormatter(padding int, v interface{}) string {
	arr, ok := v.([]interface{})
	if !ok {
		return pterm.Sprint(v)
	}
	nodes := make([]pterm.TreeNode, 0)
	for _, info := range arr {
		infoJSON, _ := json.MarshalToString(info)
		nodes = append(nodes, pterm.TreeNode{Text: infoJSON})
	}
	s, err := pterm.DefaultTree.WithRoot(pterm.TreeNode{Children: nodes}).Srender()
	if err != nil {
		return pterm.Sprint(v)
	}
	return strings.ReplaceAll("\n"+s, "\n", "\n"+strings.Repeat(" ", padding))
}
