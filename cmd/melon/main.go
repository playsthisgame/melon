package main

import (
	"runtime/debug"

	"github.com/playsthisgame/melon/internal/cli"
)

var version = "dev"

func init() {
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}
}

func main() {
	cli.Run("melon", version)
}
