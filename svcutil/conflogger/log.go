package conflogger

import (
	"os"

	"github.com/go-courier/logr"
)

var project = "unknown"

func init() {
	if projectName := os.Getenv("PROJECT_NAME"); projectName != "" {
		project = projectName
		if version := os.Getenv("PROJECT_VERSION"); version != "" {
			project += "@" + version
		}
	}
}

func SetProjectName(projectName string) {
	project = projectName
}

type Log struct {
	Level  logr.Level `env:""`
	Output OutputType `env:""`
	Format FormatType
	init   bool
}

func (l *Log) SetDefaults() {
	if l.Output == "" {
		l.Output = OutputAlways
	}

	if l.Format == "" {
		l.Format = FormatJSON
	}
}

func (l *Log) Init() {
	if !l.init {
		SetLevel(l.Level)

		if err := InstallNewPipeline(project, l.Output, l.Format); err != nil {
			panic(err)
		}
		l.init = true
	}
}
