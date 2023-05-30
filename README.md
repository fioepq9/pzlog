# pzlog

This is a pretty console writer for zerolog, powered by pterm.


## Quick Start

```go
package main

import (
  "github.com/rs/zerolog"
  "github.com/fioepq9/pzlog"
)

func main() {
  log := zerolog.New(pzlog.NewPtermWriter()).
    With().
    TimeStamp().
    Caller().
    Stack().
    Logger()
  
  log.Info().Msg("ok")
}
```

## Screenshot

![debug](./asset/debug.png)

![info](./asset/info.png)

![warn](./asset/warn.png)

![error](./asset/error.png)


## Examples

All configurations of Writer can be overridden.
### Default Config

```go
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
        s = color.ClearCode(s)
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
```

### Change Default Key Style

```go
package main

import (
  "github.com/rs/zerolog"
  "github.com/fioepq9/pzlog"
)

func main() {
  log := zerolog.New(NewPtermWriter(func(pw *PtermWriter) {
    pw.DefaultKeyStyle = func(key string, l zerolog.Level) *pterm.Style {
      return pterm.NewStyle(pterm.Bold, pterm.FgGray)
    }
  })).With().
    Timestamp().
    Stack().
    Logger()
  
  log.Info().Msg("ok")
}

```