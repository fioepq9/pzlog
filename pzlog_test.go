package pzlog

import (
	"context"
	"os"
	"testing"

	"github.com/pkg/errors"

	"github.com/rs/zerolog"
)

func FooA() error {
	return errors.New("foo")
}

func FooB() error {
	return FooA()
}

func BarA() error {
	return errors.Wrap(FooB(), "bar")
}

func BarB() error {
	return BarA()
}

func TestMain(m *testing.M) {
	log := zerolog.New(NewPtermWriter())
	zerolog.DefaultContextLogger = &log

	code := m.Run()

	os.Exit(code)
}

func TestInfo(t *testing.T) {
	log := zerolog.Ctx(context.TODO())

	logw := log.With().
		Timestamp().
		Str("foo", "bar").
		Int("number", 114514).
		Bool("flag", true).
		Logger()

	logw.Info().Msg("")

	logw.Info().Msg("test info: str, int, bool")

	logw.Info().Caller().Msg("test info: caller")

	zerolog.ErrorStackMarshaler = MarshalStack(true)

	logw.Info().Stack().Err(BarB()).Msg("test info: stack")
}
