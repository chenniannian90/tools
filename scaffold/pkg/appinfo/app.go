package appinfo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/chenniannian90/tools/scaffold/pkg/envconf"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

func New(opts ...AppCtxOption) *AppCtx {
	a := &AppCtx{}

	options := append(
		[]AppCtxOption{},
		opts...,
	)

	for i := range options {
		options[i](a)
	}

	a.cmd = &cobra.Command{}

	if feature, exists := os.LookupEnv("PROJECT_FEATURE"); exists {
		a.feature = feature
	}

	return a
}

type AppCtx struct {
	name              string
	feature           string
	version           string
	root              string
	cmd               *cobra.Command
	defaultEnvVarList []*envconf.EnvVars
	configValues      []reflect.Value
}

func (c *AppCtx) String() string {
	if c.feature != "" {
		return c.name + "--" + c.feature + "@" + c.version
	}
	return c.name + "@" + c.version
}

func (c *AppCtx) AddCommand(cmd string, fn func(args ...string), cmdOpts ...func(*cobra.Command)) {
	command := &cobra.Command{Use: cmd}

	for i := range cmdOpts {
		cmdOpts[i](command)
	}

	command.Run = func(cmd *cobra.Command, args []string) {
		fn(args...)
	}

	c.cmd.AddCommand(command)
}

func (c *AppCtx) ConfP(values ...interface{}) {
	contents, err := ioutil.ReadFile(filepath.Join(c.root, "./config/local.yml"))
	if err == nil {
		kv := map[string]string{}
		err = yaml.Unmarshal(contents, &kv)
		if err == nil {
			for k, v := range kv {
				_ = os.Setenv(k, v)
			}
		}
	}

	for i := range values {
		v := values[i]

		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Ptr {
			panic(fmt.Errorf("ConfP pass ptr for setting value"))
		}

		c.scanDefaults(rv)
		c.mustMarshal(rv)
		c.configValues = append(c.configValues, rv)

		triggerInitials(rv)
	}
}

func (c *AppCtx) Conf(values ...interface{}) {
	contents, err := ioutil.ReadFile("./config/local.yml")
	if err == nil {
		kv := map[string]string{}
		err = yaml.Unmarshal(contents, &kv)
		if err == nil {
			for k, v := range kv {
				_ = os.Setenv(k, v)
			}
		}
	}

	for i := range values {
		v := values[i]

		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Ptr {
			panic(fmt.Errorf("ConfP pass ptr for setting value"))
		}

		c.scanDefaults(rv)
		c.mustMarshal(rv)
		c.configValues = append(c.configValues, rv)

		triggerInitials(rv)
	}
}

func (c *AppCtx) Execute(fn func(args ...string), cmdOpts ...func(*cobra.Command)) {
	for i := range cmdOpts {
		cmdOpts[i](c.cmd)
	}

	c.cmd.Use = c.name
	c.cmd.Version = c.version
	c.cmd.Run = func(cmd *cobra.Command, args []string) {
		for i := range c.configValues {
			c.log(c.configValues[i])
		}
		fn(args...)
	}

	//c.AddCommand("dockerize", func(args ...string) {
	//	c.dockerize()
	//}, func(cmd *cobra.Command) {
	//	cmd.Short = "init configuration for dockerize"
	//})

	if err := c.cmd.Execute(); err != nil {
		panic(err)
	}
}

func (c *AppCtx) scanDefaults(rv reflect.Value) {
	envVars := envconf.NewEnvVars(prefix(c.name, rv.Elem().Type().Name()))

	if err := envconf.DecodeEnvVars(envVars, rv); err != nil {
		panic(err)
	}

	c.defaultEnvVarList = append(c.defaultEnvVarList, envVars)

	if err := envconf.EncodeEnvVars(envVars, rv); err != nil {
		panic(err)
	}
}

func (c *AppCtx) log(rv reflect.Value) {
	envVars := envconf.NewEnvVars(prefix(c.name, rv.Elem().Type().Name()))

	if err := envconf.EncodeEnvVars(envVars, rv); err != nil {
		panic(err)
	}

	fmt.Printf("%s", string(envVars.MaskBytes()))
}

func (c *AppCtx) mustMarshal(rv reflect.Value) {
	envVars := envconf.EnvVarsFromEnviron(prefix(c.name, rv.Elem().Type().Name()), os.Environ())

	if err := envconf.DecodeEnvVars(envVars, rv); err != nil {
		panic(err)
	}
}

type AppCtxOption = func(conf *AppCtx)

func WithName(name string) AppCtxOption {
	return func(c *AppCtx) {
		c.name = name
	}
}

func WithVersion(version string) AppCtxOption {
	return func(c *AppCtx) {
		c.version = version
	}
}

func WithMainRoot(rootDir string) AppCtxOption {
	_, filename, _, _ := runtime.Caller(1)
	parts := strings.Split(filepath.Dir(filename), "/")

	name := ""

	for i := range parts {
		if parts[i] == "cmd" && i < len(parts)-1 {
			name = parts[i+1]
		}
	}

	return func(c *AppCtx) {
		c.root = filepath.Join(filepath.Dir(filename), rootDir)
		if name != "" {
			c.name = name
		}
	}
}

func triggerInitials(rv reflect.Value) {
	rv = reflect.Indirect(rv)

	if rv.Kind() == reflect.Struct {
		for i := 0; i < rv.NumField(); i++ {
			value := rv.Field(i)

			if conf, ok := value.Interface().(interface{ Init() }); ok {
				conf.Init()
			}
		}
	} else {
		if conf, ok := rv.Interface().(interface{ Init() }); ok {
			conf.Init()
		}
	}
}
