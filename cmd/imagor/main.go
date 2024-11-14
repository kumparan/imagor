package main

import (
	"fmt"
	"github.com/cshum/imagor/config"
	"github.com/cshum/imagor/config/awsconfig"
	"github.com/cshum/imagor/config/gcloudconfig"
	"github.com/cshum/imagor/config/vipsconfig"
	"github.com/cshum/imagor/storage/base64storage"
	"os"
)

func main() {
	fmt.Println("PAS AWAL BRO")
	var server = config.CreateServer(
		os.Args[1:],
		vipsconfig.WithVips,
		awsconfig.WithAWS,
		gcloudconfig.WithGCloud,
		base64storage.WithBase64,
	)
	if server != nil {
		server.Run()
	}
}
