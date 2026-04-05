package main

import "github.com/playsthisgame/melon/internal/cli"

var version = "dev"

func main() {
	cli.Run("melon", version)
}
