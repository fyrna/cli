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
	Name     string
	Version  string
	Desc     string
	root     *node
	dynamics map[string]Dynamic // dynamic struct registry (experimental)
	plugins  []Plugin

	// behaviour hooks
	OnNotFound NotFoundHandler
	OnError    ErrorHandler

	// I/O (pluggable)
	Out io.Writer
	Err io.Writer

	// configuration
	config appConfig

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

// Dynamic allows a struct to describe itself at runtime.
type Dynamic interface {
	Metadata() Command
	Run(*Context) error
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
		Name:     name,
		root:     &node{subs: make(map[string]*node)},
		dynamics: make(map[string]Dynamic),
		OnNotFound: func(ctx *Context, s string) error {
			fmt.Fprintf(ctx.App.Err, "command %q not found\n", s)
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

func (a *App) Register(d Dynamic) *App {
	meta := d.Metadata()
	a.dynamics[meta.Name] = d
	a.add(meta.Name, &meta)
	return a
}

func (a *App) add(path string, cmd *Command) *App {
	parts := strings.Split(path, " ")
	cur, _ := a.root.get(parts[:len(parts)-1])
	name := parts[len(parts)-1]
	if _, ok := cur.subs[name]; ok {
		panic("duplicate command: " + path)
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

func (a *App) showRootHelp() error {
	if a.config.debug {
		// FIXME: root > help > rootHelp
		// i'll fix this later
		log.Println(DebugNoRootCommand)
		log.Println(DebugUsingDefaultHelp)
	}

	v := ""

	if a.Version != "" {
		v = a.Version
		fmt.Fprintf(a.Out, "%s - %s\n", a.Name, v)
	} else {
		fmt.Fprintf(a.Out, "%s\n", a.Name)
	}

	fmt.Fprintf(a.Out, "\nUsage: %s <command>\n", a.Name)

	return nil
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

// Runuwu >x<
func (a *App) Run(args []string) error {
	if a.config.debug {
		log.Println(DebugReport)
	}

	if len(args) == 0 {
		return a.showRootHelp()
	}

	n, rest := a.root.get(strings.Split(args[0], " "))

	if n.cmd == nil {
		ctx := &Context{App: a, RawArgs: rest}
		return a.OnNotFound(ctx, args[0])
	}

	if a.config.debug {
		log.Println("onii-chan, i detectu args parsedu: ", args)
	}

	return a.execute(n.cmd, args)
}

// Convenience entry point (os.Args)
// this is a simple way to parse, but
// you can use your own with app.Run() ! >x<
func (a *App) Parse() {
	if err := a.Run(os.Args[1:]); err != nil {
		ctx := &Context{App: a}
		a.OnError(ctx, err)
		os.Exit(1)
	}
}
