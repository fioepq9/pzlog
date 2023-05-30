package pzlog

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/goccy/go-json"
	"github.com/gookit/color"
	"github.com/pterm/pterm"
	"github.com/rs/zerolog"
)

type Field struct {
	Key string
	Val string
}

type Event struct {
	Timestamp string
	Level     string
	Message   string
	Fields    []Field
}

type Formatter func(interface{}) string

type PtermWriter struct {
	MaxWidth int
	Out      io.Writer

	LevelStyles map[zerolog.Level]*pterm.Style

	Tmpl *template.Template

	DefaultKeyStyle     func(string, zerolog.Level) *pterm.Style
	DefaultValFormatter func(string, zerolog.Level) Formatter

	KeyStyles     map[string]*pterm.Style
	ValFormatters map[string]Formatter

	KeyOrderFunc func(string, string) bool
}

func NewPtermWriter(options ...func(*PtermWriter)) *PtermWriter {
	pt := PtermWriter{
		MaxWidth: pterm.GetTerminalWidth(),
		Out:      os.Stdout,
		LevelStyles: map[zerolog.Level]*pterm.Style{
			zerolog.TraceLevel: pterm.NewStyle(pterm.Bold, pterm.FgCyan),
			zerolog.DebugLevel: pterm.NewStyle(pterm.Bold, pterm.FgBlue),
			zerolog.InfoLevel:  pterm.NewStyle(pterm.Bold, pterm.FgGreen),
			zerolog.WarnLevel:  pterm.NewStyle(pterm.Bold, pterm.FgYellow),
			zerolog.ErrorLevel: pterm.NewStyle(pterm.Bold, pterm.FgRed),
			zerolog.FatalLevel: pterm.NewStyle(pterm.Bold, pterm.FgRed),
			zerolog.PanicLevel: pterm.NewStyle(pterm.Bold, pterm.FgRed),
			zerolog.NoLevel:    pterm.NewStyle(pterm.Bold, pterm.FgWhite),
		},
		KeyStyles: map[string]*pterm.Style{
			zerolog.MessageFieldName:    pterm.NewStyle(pterm.Bold, pterm.FgWhite),
			zerolog.TimestampFieldName:  pterm.NewStyle(pterm.Bold, pterm.FgGray),
			zerolog.CallerFieldName:     pterm.NewStyle(pterm.Bold, pterm.FgGray),
			zerolog.ErrorFieldName:      pterm.NewStyle(pterm.Bold, pterm.FgRed),
			zerolog.ErrorStackFieldName: pterm.NewStyle(pterm.Bold, pterm.FgRed),
		},
		ValFormatters: map[string]Formatter{},
		KeyOrderFunc: func(k1, k2 string) bool {
			score := func(s string) string {
				if s == zerolog.TimestampFieldName {
					return string([]byte{0, 0})
				}
				if s == zerolog.CallerFieldName {
					return string([]byte{0, 1})
				}
				if s == zerolog.ErrorFieldName {
					return string([]byte{math.MaxUint8, 0})
				}
				if s == zerolog.ErrorStackFieldName {
					return string([]byte{math.MaxUint8, 1})
				}
				return s
			}
			return score(k1) < score(k2)
		},
	}

	tmpl := `{{ .Timestamp }} {{ .Level }}  {{ .Message }}
{{- range $i, $field := .Fields }}
{{ space (totalLength 1 $.Timestamp $.Level) }}{{if (last $i $.Fields )}}└{{else}}├{{ end }} {{ .Key }}: {{ .Val }}
{{- end }}
`
	t, err := template.New("event").
		Funcs(template.FuncMap{
			"space": func(n int) string {
				return strings.Repeat(" ", n)
			},
			"totalLength": func(n int, s ...string) int {
				return len(color.ClearCode(strings.Join(s, ""))) - n
			},
			"last": func(x int, a interface{}) bool {
				return x == reflect.ValueOf(a).Len()-1
			},
		}).
		Parse(tmpl)
	if err != nil {
		panic(fmt.Errorf("cannot parse template: %s", err))
	}
	pt.Tmpl = t

	pt.DefaultKeyStyle = func(_ string, lvl zerolog.Level) *pterm.Style {
		return pt.LevelStyles[lvl]
	}

	pt.DefaultValFormatter = func(key string, lvl zerolog.Level) Formatter {
		return func(v interface{}) string {
			return pterm.Sprint(v)
		}
	}

	for _, option := range options {
		option(&pt)
	}

	return &pt
}

func (pt *PtermWriter) Write(p []byte) (n int, err error) {
	return pt.Out.Write(p)
}

func (pt *PtermWriter) WriteLevel(lvl zerolog.Level, p []byte) (n int, err error) {
	var evt map[string]interface{}
	d := json.NewDecoder(bytes.NewReader(p))
	d.UseNumber()
	err = d.Decode(&evt)
	if err != nil {
		return n, fmt.Errorf("cannot decode event: %s", err)
	}

	var event Event
	if ts, ok := evt[zerolog.TimestampFieldName]; ok {
		event.Timestamp = pt.KeyStyles[zerolog.TimestampFieldName].Sprint(ts)
	}
	event.Level = pt.LevelStyles[lvl].Sprint(lvl)
	if msg, ok := evt[zerolog.MessageFieldName]; ok {
		event.Message = pt.KeyStyles[zerolog.MessageFieldName].Sprint(msg)
	}
	event.Fields = make([]Field, 0, len(evt))
	for k, v := range evt {
		if k == zerolog.TimestampFieldName ||
			k == zerolog.LevelFieldName ||
			k == zerolog.MessageFieldName {
			continue
		}
		var key string
		if style, ok := pt.KeyStyles[k]; ok {
			key = style.Sprint(k)
		} else {
			key = pt.DefaultKeyStyle(k, lvl).Sprint(k)
		}
		var val string
		if fn, ok := pt.ValFormatters[k]; ok {
			val = fn(v)
		} else {
			val = pt.DefaultValFormatter(k, lvl)(v)
		}
		event.Fields = append(event.Fields, Field{Key: key, Val: val})
	}
	var buf bytes.Buffer
	err = pt.Tmpl.Execute(&buf, event)
	if err != nil {
		return n, fmt.Errorf("cannot execute template: %s", err)
	}
	return pt.Out.Write(buf.Bytes())
}
