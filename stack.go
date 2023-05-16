package pzlog

import (
	"math"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

var (
	StackSourceFileName     = "source"
	StackSourceLineName     = "line"
	StackSourceFunctionName = "func"
)

type state struct {
	flag [math.MaxUint8]bool
	b    []byte
}

// Write implement fmt.Formatter interface.
func (s *state) Write(b []byte) (n int, err error) {
	s.b = b
	return len(b), nil
}

// Width implement fmt.Formatter interface.
func (s *state) Width() (wid int, ok bool) {
	return 0, false
}

// Precision implement fmt.Formatter interface.
func (s *state) Precision() (prec int, ok bool) {
	return 0, false
}

// Flag implement fmt.Formatter interface.
func (s *state) Flag(c int) bool {
	return s.flag[c]
}

func frameField(f errors.Frame, c rune) string {
	s := &state{}
	s.flag['+'] = true
	f.Format(s, c)
	return string(s.b)
}

// MarshalStack implements pkg/errors stack trace marshaling.
//
//	zerolog.ErrorStackMarshaler = MarshalStack()
func MarshalStack(simple bool) func(err error) interface{} {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}
	_, file, _, _ := runtime.Caller(1)
	wd := filepath.Dir(file)
	return func(err error) interface{} {
		out := make([]map[string]string, 0)
		var rootErr stackTracer
		for ; err != nil; err = errors.Unwrap(err) {
			sterr, ok := err.(stackTracer)
			if !ok {
				continue
			}
			rootErr = sterr
		}
		st := rootErr.StackTrace()
		for _, frame := range st {
			srcPath := frameField(frame, 's')
			if simple {
				if !strings.HasPrefix(srcPath, wd) {
					continue
				}
				srcPath = "." + strings.TrimPrefix(srcPath, wd)
			}
			out = append(out, map[string]string{
				StackSourceFileName:     srcPath,
				StackSourceLineName:     frameField(frame, 'd'),
				StackSourceFunctionName: frameField(frame, 'n'),
			})
		}
		return out
	}
}
