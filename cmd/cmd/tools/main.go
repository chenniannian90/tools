package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chenniannian90/tools/cmd/cmd/tools/gen"
	"github.com/chenniannian90/tools/cmd/cmd/tools/hook"
	"github.com/chenniannian90/tools/cmd/version"
	"github.com/go-courier/logr"
	"github.com/spf13/cobra"
)

var verbose = false

var cmdRoot = &cobra.Command{
	Use:     "tools",
	Version: version.Version,
}

func init() {
	cmdRoot.PersistentFlags().BoolVarP(&verbose, "verbose", "v", verbose, "")

	cmdRoot.AddCommand(gen.CmdGen)
	cmdRoot.AddCommand(hook.CmdHook)
}

func main() {
	ctx := logr.WithLogger(context.Background(), logr.StdLogger())

	if err := cmdRoot.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
