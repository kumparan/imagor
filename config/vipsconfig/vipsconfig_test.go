package vipsconfig

import (
	"github.com/kumparan/imagor"
	"github.com/kumparan/imagor/config"
	"github.com/kumparan/imagor/processor/vipsprocessor"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithVips(t *testing.T) {
	srv := config.CreateServer([]string{
		"-vips-max-animation-frames", "167",
		"-vips-disable-filters", "blur,watermark,rgb",
	}, WithVips)
	app := srv.App.(*imagor.Imagor)
	processor := app.Processors[0].(*vipsprocessor.Processor)
	assert.Equal(t, 167, processor.MaxAnimationFrames)
	assert.Equal(t, []string{"blur", "watermark", "rgb"}, processor.DisableFilters)
}
