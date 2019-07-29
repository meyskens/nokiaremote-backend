package main

import "gopkg.in/src-d/go-cli.v0"

var (
	name    = "nokiaremote"
	version = "undefined"
	build   = "undefined"
)

var app = cli.New(name, version, build, "The backend system for NokiaRemote")

func main() {
	app.RunMain()
}
