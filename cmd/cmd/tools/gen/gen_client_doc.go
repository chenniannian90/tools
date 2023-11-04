package gen

//import (
//	"fmt"
//	"net/url"
//	"os"
//	"path"
//
//	"tools/cmd/internal/generate"
//	"tools/openapi2word"
//	"github.com/go-courier/packagesx"
//
//	"github.com/spf13/cobra"
//)
//
//var (
//	cmdGenClientDocFlagFile     string
//	cmdGenClientDocFlagSpecURL  string
//	cmdGenClientDocHeadingLevel int
//)
//
//func init() {
//
//	cmdGenClientDoc.Flags().
//		StringVarP(&cmdGenClientDocFlagSpecURL, "spec-url", "", "", "client-doc spec url")
//	cmdGenClientDoc.Flags().
//		StringVarP(&cmdGenClientDocFlagFile, "file", "", "", "client-doc spec file")
//	cmdGenClientDoc.Flags().
//		IntVarP(&cmdGenClientDocHeadingLevel, "heading-level", "", 3, "word doc heading level, from level 1 to 8, default 3")
//
//	CmdGen.AddCommand(cmdGenClientDoc)
//}
//
//var cmdGenClientDoc = &cobra.Command{
//	Use:     "client-doc",
//	Example: "client-doc demo",
//	Short:   "generate client word document by open api",
//	Run: func(cmd *cobra.Command, args []string) {
//		if len(args) <= 0 {
//			panic(fmt.Errorf("need service name"))
//		}
//		u := &url.URL{}
//
//		if cmdGenClientDocFlagFile != "" {
//			u.Scheme = "file"
//			cwd, _ := os.Getwd()
//			u.Path = path.Join(cwd, cmdGenClientDocFlagFile)
//		}
//
//		if cmdGenClientDocFlagSpecURL != "" {
//			uri, err := url.Parse(cmdGenClientDocFlagSpecURL)
//			if err != nil {
//				panic(err)
//			}
//			u = uri
//		}
//		if cmdGenClientDocHeadingLevel > 8 || cmdGenClientDocHeadingLevel < 1 {
//			panic(fmt.Errorf("word doc heading level, from level 1 to 8"))
//		}
//
//		fmt.Println("HeadingLevel：", cmdGenClientDocHeadingLevel)
//
//		generate.RunGenerator(func(pkg *packagesx.Package) generate.Generator {
//			g := openapi2word.NewGenerateOpenAPIDoc(args[0], u, cmdGenClientDocHeadingLevel)
//			g.Load()
//			return g
//		})
//	},
//}
