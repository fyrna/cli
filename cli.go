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
	root     *node
	dynamics map[string]Dynamic // dynamic struct registry (experimental, I will probably delete it)
	plugins  []Plugin

	// behaviour hooks
	OnNotFound NotFoundHandler
	OnError    ErrorHandler

	// I/O (pluggable)
	Out io.Writer
	Err io.Writer

	// configuration
	Config AppConfig

	flatCmds map[string]*Command
}

type AppConfig struct {
	Debug bool
}

// Context carries data through the call chain.
type Context struct {
	App     *App
	Cmd     *Command
	RawArgs []string
	Store   map[string]any

	// TODO: implement our own, for example:
	//     app := cli.New()
	//
	//     -- global flag --
	//     app.Flags(cli.StringFlag())
	//     -- or --
	//     app.StringFlag() -- direct --
	//
	//     cmd := app.Commands("", cli.Short(""), func("") error {})
	//     -- command flag : there's 2 type: strict & inherit
	//     -- * strict means only for "x" command
	//     -- * inherit means subcommand for "x" can you it too
	//     -- for default we set to strict, user can define it as "configurable"
	//     cmd.Flags(StringFlag())
	//     -- or --
	//     cmd.StringFlag() -- direct --
	Flags *flag.FlagSet
}

func (c *Context) Args() []string {
	return c.Flags.Args()
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

// --
// Internal tree node
// --

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

// --
// Constructor
// --

func New(name string) *App {
	return &App{
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
}

// Fluent API for the three styles

// 1. Simple style
// func (a *App) Command(path string, fn func(*Context) error) *App {
// return a.add(path, &Command{Action: fn})
// }

// 2. Metadata via option
type Option func(*Command)

func Short(s string) Option                 { return func(c *Command) { c.Short = s } }
func Long(s string) Option                  { return func(c *Command) { c.Long = s } }
func Alias(a ...string) Option              { return func(c *Command) { c.Aliases = a } }
func Usage(u string) Option                 { return func(c *Command) { c.Usage = u } }
func Category(cat string) Option            { return func(c *Command) { c.Category = cat } }
func Before(fn func(*Context) error) Option { return func(c *Command) { c.Before = fn } }
func After(fn func(*Context) error) Option  { return func(c *Command) { c.After = fn } }
func Flags(fs *flag.FlagSet) Option         { return func(c *Command) { c.Flags = fs } }

func (a *App) Command(path string, action func(*Context) error, opts ...Option) *App {
	cmd := &Command{Action: action}
	for _, o := range opts {
		o(cmd)
	}
	return a.add(path, cmd)
}

// 3. Dynamic struct
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

// -- Plugin & I/O helpers --
func (a *App) Use(p ...Plugin) *App {
	for _, pl := range p {
		if err := pl.Install(a); err != nil {
			panic(err)
		}
	}
	return a
}

// -- Run program --
func (a *App) Run(args []string) error {

	if a.Config.Debug {
		log.Println("report bug: https://github.com/fyrna/cli/issues")
	}

	if len(args) == 0 {
		if a.Config.Debug {
			log.Println("no help command provided, using default help...")
		}
		return a.showRootHelp()
	}
	n, rest := a.root.get(strings.Split(args[0], " "))
	if n.cmd == nil {
		ctx := &Context{App: a, RawArgs: rest}
		return a.OnNotFound(ctx, args[0])
	}

	if a.Config.Debug && len(args) >= 1 {
		log.Println("we detect arg(s) parsed: ", args)
	}

	return a.execute(n.cmd, args)
}

func (a *App) execute(c *Command, args []string) (err error) {
	fs := c.Flags
	if fs == nil {
		fs = flag.NewFlagSet(c.Name, flag.ContinueOnError)
	}
	if err := fs.Parse(args); err != nil {
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
		return errors.New("no action")
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

func (a *App) showRootHelp() error {
	fmt.Fprintf(a.Out, `Usage: %s <command>\n`, a.Name)

	return nil
}

// --
// Convenience entry point (os.Args)
// this is a simple way to parse, but
// you can use your own! app.Run() ! >x<
// --

func (a *App) Parse() {
	if err := a.Run(os.Args[1:]); err != nil {
		ctx := &Context{App: a}
		a.OnError(ctx, err)
		os.Exit(1)
	}
}
