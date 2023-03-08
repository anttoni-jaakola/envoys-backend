package main

import (
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/server"
	"runtime"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {

	// The purpose of this code is to initiate a master instance of the server with a specific context. The context defines
	// the environment and settings that the server should use when processing requests. This allows the server to customize
	// its behavior for a given context.
	server.Master(&assets.Context{})
}
