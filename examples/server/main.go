package main

import (
	"github.com/kumparan/imagor"
	"github.com/kumparan/imagor/imagorpath"
	"github.com/kumparan/imagor/loader/httploader"
	"github.com/kumparan/imagor/server"
	"github.com/kumparan/imagor/storage/filestorage"
	"github.com/kumparan/imagor/vips"
	"go.uber.org/zap"
)

func main() {
	logger := zap.Must(zap.NewProduction())

	// create and run imagor server programmatically
	server.New(
		imagor.New(
			imagor.WithLogger(logger),
			imagor.WithUnsafe(true),
			imagor.WithProcessors(vips.NewProcessor()),
			imagor.WithLoaders(httploader.New()),
			imagor.WithStorages(filestorage.New("./")),
			imagor.WithResultStorages(filestorage.New("./")),
			imagor.WithResultStoragePathStyle(imagorpath.SuffixResultStorageHasher),
		),
		server.WithPort(8000),
		server.WithLogger(logger),
	).Run()
}
