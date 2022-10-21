package main

import (
	"github.com/alecthomas/kong"
	"github.com/emm035/qrcaas/pkg/service"
)

type Cli struct {
	Config service.Config `embed:""`
}

func main() {
	ctx := kong.Parse(new(Cli))
	ctx.FatalIfErrorf(ctx.Run())
}

func (cli *Cli) Run() error {
	return service.New(cli.Config).Start(":8080")
}
