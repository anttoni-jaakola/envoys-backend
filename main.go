package main

import (
	"os"
	"runtime"

	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/server"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {

	// The purpose of this code is to get the current working directory of the operating system and store it in the variable
	// dir. The os.Getwd() function is used to do this and it returns a string representing the path of the current working
	// directory and an error value. If there is an error, it will be handled by the if statement which will cause the program to panic.
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// The purpose of this code is to initiate a master instance of the server with a specific context. The context defines
	// the environment and settings that the server should use when processing requests. This allows the server to customize
	// its behavior for a given context.
	server.Register(&assets.Context{
		StoragePath: dir,
	})
}
