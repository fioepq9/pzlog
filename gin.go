package pzlog

import (
	"io"

	"github.com/rs/zerolog"
)

type GinDefaultWriter struct {
	log *zerolog.Logger
	lvl zerolog.Level
}

func (w *GinDefaultWriter) Write(p []byte) (n int, err error) {
	w.log.WithLevel(w.lvl).Msg(string(p))
	return len(p), nil
}

func NewGinDefaultWriter(log *zerolog.Logger, lvl zerolog.Level) io.Writer {
	return &GinDefaultWriter{
		log: log,
		lvl: lvl,
	}
}
