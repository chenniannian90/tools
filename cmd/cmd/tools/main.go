package main

import (
	"context"
	"fmt"
	"github.com/go-courier/httptransport/openapi/generator"
	"github.com/go-courier/packagesx"
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

	// pwd := "/Users/mac-new/go/src/github.com/chenniannian90/chan-go/cmd/chan-go"
	pwd := "/Users/mac-new/go/src/ext-gitlab.denglin.com/ci-tool/test-manager/cmd/test-manager"
	pkg, _ := packagesx.Load(pwd)

	g := generator.NewOpenAPIGenerator(pkg)
	g.Scan(context.Background())

	ctx := logr.WithLogger(context.Background(), logr.StdLogger())

	if err := cmdRoot.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
