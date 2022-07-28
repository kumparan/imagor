package config

import (
	"flag"
	"github.com/cshum/imagor"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestImagorFuncs(t *testing.T) {
	fs := flag.NewFlagSet("imagor", flag.ExitOnError)
	nopLogger := zap.NewNop()
	var seq []int
	NewImagor(fs, func() (logger *zap.Logger, isDebug bool) {
		seq = append(seq, 4)
		return nopLogger, true
	}, func(fs *flag.FlagSet, cb Callback) imagor.Option {
		seq = append(seq, 3)
		logger, isDebug := cb()
		assert.Equal(t, nopLogger, logger)
		assert.True(t, isDebug)
		seq = append(seq, 5)
		return func(app *imagor.Imagor) {
			seq = append(seq, 8)
		}
	}, func(fs *flag.FlagSet, cb Callback) imagor.Option {
		seq = append(seq, 2)
		logger, isDebug := cb()
		assert.Equal(t, nopLogger, logger)
		assert.True(t, isDebug)
		seq = append(seq, 6)
		return func(app *imagor.Imagor) {
			seq = append(seq, 9)
		}
	}, func(fs *flag.FlagSet, cb Callback) imagor.Option {
		seq = append(seq, 1)
		logger, isDebug := cb()
		assert.Equal(t, nopLogger, logger)
		assert.True(t, isDebug)
		seq = append(seq, 7)
		return func(app *imagor.Imagor) {
			seq = append(seq, 10)
		}
	})
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, seq)
}

func TestImagorFuncsNil(t *testing.T) {
	fs := flag.NewFlagSet("imagor", flag.ExitOnError)
	nopLogger := zap.NewNop()
	var seq []int
	NewImagor(fs, func() (logger *zap.Logger, isDebug bool) {
		seq = append(seq, 4)
		return nopLogger, true
	}, func(fs *flag.FlagSet, cb Callback) imagor.Option {
		seq = append(seq, 3)
		logger, isDebug := cb()
		assert.Equal(t, nopLogger, logger)
		assert.True(t, isDebug)
		seq = append(seq, 5)
		return func(app *imagor.Imagor) {
			seq = append(seq, 7)
		}
	}, nil, func(fs *flag.FlagSet, cb Callback) imagor.Option {
		seq = append(seq, 2)
		return func(app *imagor.Imagor) {
			seq = append(seq, 8)
		}
	}, nil, func(fs *flag.FlagSet, cb Callback) imagor.Option {
		seq = append(seq, 1)
		logger, isDebug := cb()
		assert.Equal(t, nopLogger, logger)
		assert.True(t, isDebug)
		seq = append(seq, 6)
		return func(app *imagor.Imagor) {
			seq = append(seq, 9)
		}
	})
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}, seq)
}