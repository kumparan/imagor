package main

import (
	"github.com/kumparan/imagor/config"
	"github.com/kumparan/imagor/config/awsconfig"
	"github.com/kumparan/imagor/config/gcloudconfig"
	"github.com/kumparan/imagor/config/vipsconfig"
	"os"
)

func main() {
	var server = config.CreateServer(
		os.Args[1:],
		vipsconfig.WithVips,
		awsconfig.WithAWS,
		gcloudconfig.WithGCloud,
	)
	if server != nil {
		server.Run()
	}
}
