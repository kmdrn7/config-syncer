package main

import (
	"config-syncer/cmd"
)

// TODO: wrap in cobra style
// TODO: implement sync using configurable YAML config file instead of using src-secret and dest-secret
// TODO: prepare helm charts

func main() {
	cmd.Execute()
}
