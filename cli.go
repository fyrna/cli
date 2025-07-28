// fyrna/cli is a tiny, flexible command-line micro-framework.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// App is the root CLI application.
type App struct {
	Name    string
	Version string
	Desc    string

	// behaviour hooks
	OnNotFound NotFoundHandler
	OnError    ErrorHandler

	// I/O (pluggable)
	Out io.Writer
	Err io.Writer

	// configuration
	config appConfig

	// internal use
	root     *node
	plugins  []Plugin
	flatCmds map[string]*Command
}

type appConfig struct {
	debug bool
}

// Command is the canonical representation of a runnable thing.
type Command struct {
	Name     string
	Aliases  []string
	Usage    string
	Short    string
	Long     string
	Category string

	Before func(*Context) error
	Action func(*Context) error
	After  func(*Context) error

	Flags *flag.FlagSet
}

// Plugin is the extension point.
type Plugin interface{ Install(*App) error }

// Handlers
type NotFoundHandler func(*Context, string) error
type ErrorHandler func(*Context, error) error

// handler for print all commands
func (a *App) Commands() map[string]*Command { return a.flatCmds }

// Internal tree node
type node struct {
	cmd  *Command
	subs map[string]*node
}

func (n *node) get(parts []string) (*node, []string) {
	cur := n
	for i, p := range parts {
		next, ok := cur.subs[p]
		if !ok {
			return cur, parts[i:]
		}
		cur = next
	}
	return cur, nil
}

// Constructor
func New(name string, opts ...ConfigOption) *App {
	app := &App{
		Name: name,
		root: &node{subs: make(map[string]*node)},
		OnNotFound: func(ctx *Context, s string) error {
			fmt.Fprintf(ctx.App.Err, errCommandNotFound, s)
			return nil
		},
		OnError: func(ctx *Context, err error) error {
			fmt.Fprintln(ctx.App.Err, err)
			return err
		},
		Out: os.Stdout,
		Err: os.Stderr,
	}

	for _, o := range opts {
		o(app)
	}

	app.Use(printAppVersion{})

	return app
}

func (a *App) Command(path string, actionOrOps ...any) *App {
	cmd := &Command{Name: path}

	if len(actionOrOps) == 1 {
		if fn, ok := actionOrOps[0].(func(*Context) error); ok {
			cmd.Action = fn
			return a.add(path, cmd)
		}
	}

	for _, opt := range actionOrOps {
		if o, ok := opt.(CommandOption); ok {
			o(cmd)
		}
	}

	return a.add(path, cmd)
}

func isBuiltin(name string) bool {
	switch name {
	case "version", "help", "verbose":
		return true
	}
	return false
}

func (a *App) add(path string, cmd *Command) *App {
	// root override
	if path == rootCommandName {
		a.root.cmd = cmd
		return a
	}

	parts := strings.Split(path, " ")
	cur, _ := a.root.get(parts[:len(parts)-1])
	name := parts[len(parts)-1]

	if _, ok := cur.subs[name]; ok {
		if isBuiltin(name) {
			delete(cur.subs, name)
		} else {
			panic(errDuplicateCommand + path)
		}
	}

	cmd.Name = name
	cur.subs[name] = &node{cmd: cmd, subs: make(map[string]*node)}
	return a
}

// Installing plugins
func (a *App) Use(p ...Plugin) *App {
	for _, pl := range p {
		if err := pl.Install(a); err != nil {
			panic(err)
		}
	}
	return a
}

func (a *App) execute(c *Command, args []string) (err error) {
	fs := c.Flags
	if fs == nil {
		fs = flag.NewFlagSet(c.Name, flag.ContinueOnError)
	}

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	ctx := &Context{
		App:     a,
		Cmd:     c,
		RawArgs: args,
		Flags:   fs,
		Store:   make(map[string]any),
	}

	if c.Before != nil {
		if err = c.Before(ctx); err != nil {
			return err
		}
	}

	if c.Action == nil {
		return errors.New("onii-chan.. no action defined")
	}

	defer func() {
		if c.After != nil {
			if e := c.After(ctx); e != nil && err == nil {
				err = e
			}
		}
	}()

	return c.Action(ctx)
}

// parser
func (a *App) Parse(args []string) error {
	if a.config.debug {
		log.Println(debugReport)
	}

	if len(args) == 0 {
		if a.config.debug {
			log.Println(debugNoRootCommand)
			log.Println(debugUsingDefaultHelp)
		}

		// 1) root command
		if a.root.cmd != nil {
			return a.execute(a.root.cmd, []string{rootCommandName})
		}

		// 2) help command
		h, ok := a.root.subs["help"]
		if ok && h.cmd != nil {
			return a.execute(h.cmd, []string{"help"})
		}

		// 3) default
		return a.ShowRootHelp()
	}

	n, rest := a.root.get(args)

	if n.cmd == nil {
		ctx := &Context{App: a, RawArgs: rest}
		return a.OnNotFound(ctx, args[0])
	}

	if a.config.debug {
		log.Printf(debugArgsParsed, args)
	}

	return a.execute(n.cmd, args)
}

// Builtin runner
func (a *App) Run() {
	if err := a.Parse(os.Args[1:]); err != nil {
		ctx := &Context{App: a}
		a.OnError(ctx, err)
		os.Exit(1)
	}
}
