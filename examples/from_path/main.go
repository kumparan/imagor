package main

import (
	"context"
	"github.com/kumparan/imagor"
	"github.com/kumparan/imagor/imagorpath"
	"github.com/kumparan/imagor/loader/httploader"
	"github.com/kumparan/imagor/vips"
	"io"
	"os"
)

func main() {
	app := imagor.New(
		imagor.WithLoaders(httploader.New()),
		imagor.WithProcessors(vips.NewProcessor()),
	)
	ctx := context.Background()
	if err := app.Startup(ctx); err != nil {
		panic(err)
	}
	defer app.Shutdown(ctx)
	// serve via image path
	out, err := app.Serve(ctx, imagorpath.Params{
		Image:  "https://raw.githubusercontent.com/cshum/imagor/master/testdata/gopher.png",
		Width:  500,
		Height: 500,
		Smart:  true,
		Filters: []imagorpath.Filter{
			{"fill", "white"},
			{"format", "jpg"},
		},
	})
	if err != nil {
		panic(err)
	}
	reader, _, err := out.NewReader()
	if err != nil {
		panic(err)
	}
	defer reader.Close()
	file, err := os.Create("gopher.jpg")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if _, err := io.Copy(file, reader); err != nil {
		panic(err)
	}
}
