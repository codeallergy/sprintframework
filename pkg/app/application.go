/*
 * Copyright (c) 2022-2023 Zander Schwid & Co. LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 */

package app

import (
	"flag"
	"fmt"
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprint"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Option configures badger using the functional options paradigm
// popularized by Rob Pike and Dave Cheney. If you're unfamiliar with this style,
// see https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html and
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis.
type Option interface {
	apply(sprint.Application)
}

// OptionFunc implements Option interface.
type optionFunc func(sprint.Application)

// apply the configuration to the provided config.
func (fn optionFunc) apply(a sprint.Application) {
	fn(a)
}

// option that do nothing
func WithNope() Option {
	return optionFunc(func(sprint.Application) {
	})
}

// option that adds name to application
func WithName(name string) Option {
	return optionFunc(func(a sprint.Application) {
		if app, ok := a.(*application); ok {
			app.applicationName = name
		}
	})
}

// option that adds version to application
func WithVersion(version string) Option {
	return optionFunc(func(a sprint.Application) {
		if app, ok := a.(*application); ok {
			app.applicationVersion = version
		}
	})
}

func WithBuild(build string) Option {
	return optionFunc(func(a sprint.Application) {
		if app, ok := a.(*application); ok {
			app.applicationBuild = build
		}
	})
}

func Beans(beans ...interface{}) Option {
	return optionFunc(func(a sprint.Application) {
		a.AppendBeans(beans...)
	})
}

func Core(scanner sprint.CoreScanner) Option {
	return optionFunc(func(a sprint.Application) {
		a.AppendBeans(scanner)
	})
}

func Server(scanner sprint.ServerScanner) Option {
	return optionFunc(func(a sprint.Application) {
		a.AppendBeans(scanner)
	})
}

func Client(scanner sprint.ClientScanner) Option {
	return optionFunc(func(a sprint.Application) {
		a.AppendBeans(scanner)
	})
}

type application struct {

	beans []interface{}

	applicationName    string
	applicationVersion string
	applicationBuild   string
	applicationProfile string

	applicationErr   atomic.Error

	executable     string
	executableDir  string
	applicationDir string

	devMode       bool

	shuttingDown  atomic.Bool
	shutdownCh    chan struct{}   // sends only close channel event
	restarting    atomic.Bool
	shutdownOnce  sync.Once

}

type applicationDep struct {
	ApplicationFlags sprint.ApplicationFlags   `inject`
	FlagSet          *flag.FlagSet            `inject`
	Commands         map[string]sprint.Command `inject`
}

func Application(name string, options ... Option) sprint.Application {
	t := &application{
		applicationName:  name,
		shutdownCh:   make(chan struct{}),
	}
	t.applicationErr.Store(nil)

	for _, opt := range options {
		opt.apply(t)
	}

	t.beans = append(t.beans, t)
	return t
}

func (t *application) BeanName() string {
	return "application"
}

func (t *application) GetStats(cb func(name, value string) bool) error {
	cb("name", t.applicationName)
	cb("executable", t.executable)
	cb("home", t.applicationDir)
	cb("version", t.applicationVersion)
	cb("build", t.applicationBuild)
	cb("profile", t.applicationProfile)
	return nil
}

func (t *application) PostConstruct() (err error) {

	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			case string:
				err = errors.New(v)
			default:
				err = errors.Errorf("%v", v)
			}
		}
	}()

	t.executable = os.Args[0]
	t.executableDir, err = filepath.Abs(filepath.Dir(t.executable))
	if err != nil {
		return err
	}
	t.executable = filepath.Base(t.executable)
	if filepath.Base(t.executableDir) == "bin" {
		t.applicationDir, err = filepath.Abs(filepath.Dir(t.executableDir))
		if err != nil {
			return err
		}
	} else {
		t.applicationDir, err = filepath.Abs(t.executableDir)
		if err != nil {
			return err
		}
	}
	envName := strings.ToUpper(fmt.Sprintf("%s_%s", t.applicationName, "profile"))
	t.applicationProfile = strings.ToLower(os.Getenv(envName))
	t.devMode = t.applicationProfile == "dev"
	return nil
}

func (t *application) AppendBeans(scan ...interface{}) {
	t.beans = append(t.beans, scan...)
}

func (t *application) Name() string {
	return t.applicationName
}

func (t *application) Version() string {
	return t.applicationVersion
}

func (t *application) Build() string {
	return t.applicationBuild
}

func (t *application) Profile() string {
	return t.applicationProfile
}

func (t *application) IsDev() bool {
	return t.devMode
}

func (t *application) Executable() string {
	return t.executable
}

func (t *application) ApplicationDir() string {
	return t.applicationDir
}

func (t *application) Active() bool {
	return !t.shuttingDown.Load()
}

func (t *application) Shutdown(restart bool) {
	t.restarting.Store(restart)
	t.shuttingDown.Store(true)
	t.shutdownOnce.Do(func() {
		close(t.shutdownCh)
	})
	t.applicationErr.Store(errors.New("application shutting down"))
}

func (t *application) Restarting() bool {
	return t.restarting.Load()
}

func (t *application) Deadline() (deadline time.Time, ok bool) {
	return time.Now(), false
}

func (t *application) Value(key interface{}) interface{} {
	return nil
}

func (t *application) Done() <-chan struct{} {
	return t.shutdownCh
}

func (t *application) Err() error {
	return t.applicationErr.Load()
}

func (t *application) Run(args []string) (err error) {

	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			case string:
				err = errors.New(v)
			default:
				err = errors.Errorf("%v", v)
			}
		}
	}()

	rand.Seed(time.Now().UnixNano())

	args = preprocessArgs(args)

	dep := &applicationDep{}
	propertyFile := &glue.PropertySource{ Path: fmt.Sprintf("resources:%s.yml", t.applicationName) }
	propertyMap := &glue.PropertySource{ Map: map[string]interface{} {
		"application": map[string]interface{} {
			"name": t.applicationName,
			"version": t.applicationVersion,
			"build": t.applicationBuild,
			"profile": t.applicationProfile,
			"perm": DefaultFileModes,
			"autoupdate": false,
		},
	}}
	t.AppendBeans(dep, propertyFile, propertyMap, SystemEnvironmentPropertyResolver(t.applicationName, 10))

	ctx, err := glue.New(t.beans)
	if err != nil {
		return err
	}
	defer ctx.Close()
	
	if err := dep.FlagSet.Parse(args); err != nil {
		return err
	}
	args = dep.FlagSet.Args()

	if len(args) >= 1 {

		cmd := args[0]

		if inst, ok := dep.Commands[cmd]; ok {
			return inst.Run(args[1:])
		} else {
			fmt.Printf("Invalid command: %s\n", cmd)
			t.printUsage(dep)
			return nil
		}

	} else if inst, ok := dep.Commands["default"]; ok {
		return inst.Run(args[1:])
	} else {
		t.printUsage(dep)
		return nil
	}
}

func preprocessArgs(args []string) []string {

	if len(args) == 1 && (args[0] == "-h" || args[0] == "-help" || args[0] == "--help") {
		return []string{"help"}
	}

	if len(args) == 1 && (args[0] == "-v" || args[0] == "-version" || args[0] == "--version") {
		return []string{"version"}
	}

	return args
}

func (t *application) printUsage(dep *applicationDep) {

	fmt.Printf("Usage: %s [command]\n", t.executable)

	for _, command := range dep.Commands {
		fmt.Printf("    %s - %s\n", command.BeanName(), command.Desc())
	}

	fmt.Println("Flags:")
	dep.FlagSet.PrintDefaults()

}

