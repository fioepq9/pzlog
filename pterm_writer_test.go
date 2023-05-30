package pzlog

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestMain(m *testing.M) {
	log := zerolog.New(NewPtermWriter()).With().
		Timestamp().
		Stack().
		Logger()
	zerolog.TimeFieldFormat = time.DateTime
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.DefaultContextLogger = &log

	code := m.Run()

	os.Exit(code)
}

func TestDebug(t *testing.T) {
	log := zerolog.Ctx(context.TODO())

	log.Debug().
		Str("foo", "bar").
		Int("number", 114514).
		Bool("flag", true).
		Dur("duration", time.Second).
		Time("time", time.Now()).
		Msg("debug case 1")

	log.Debug().
		Str("when were crayons invented", "1885").
		Str("when was the first spaceflight", "1957").
		Str("when did the first person land on the moon", "1969").
		Msg("debug case 2")
}

func TestInfo(t *testing.T) {
	log := zerolog.Ctx(context.TODO())

	log.Info().
		Str("foo", "bar").
		Int("number", 114514).
		Bool("flag", true).
		Dur("duration", time.Second).
		Time("time", time.Now()).
		Msg("info case 1")

	log.Info().
		Int("what is the meaning of life", 42).
		Int("what is the meaning of death", 0).
		Msg("info case 2")
}

func TestWarn(t *testing.T) {
	log := zerolog.Ctx(context.TODO())

	log.Warn().
		Str("foo", "bar").
		Int("number", 114514).
		Bool("flag", true).
		Dur("duration", time.Second).
		Time("time", time.Now()).
		Msg("warn case 1")

	log.Warn().
		Str("Oh no, I see an error coming to us!", "Oh no, I see an error coming to us!").
		Msg("warn case 2")
}

func TestError(t *testing.T) {
	log := zerolog.Ctx(context.TODO())

	log.Error().
		Str("foo", "bar").
		Int("number", 114514).
		Bool("flag", true).
		Dur("duration", time.Second).
		Time("time", time.Now()).
		Msg("error case 1")

	log.Error().
		Str("something bad happened", "something bad happened").
		Int("error code", 114514).
		Msg("error case 2")
}
