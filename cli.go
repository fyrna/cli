// fyrna/cli is a tiny, flexible command-line micro-framework.
package cli

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// internal sentinel for root override
const rootCommandPath = ""

// App represents your command line application.
// Create on using New() and configure it with "settings", commands, and flags
//
//	app := cli.New("app")
//	app.Command("serve", func(c *cli.Context) error { ... })
type App struct {
	Name    string
	Version string
	Desc    string

	// behaviour hooks
	OnNotFound NotFoundHandler
	OnError    ErrorHandler

	// I/O streams used by the framework and user handlers.
	// Default are os.Stdout and os.Stderr respectively.
	Out io.Writer // normal command output
	Err io.Writer // error messages output

	// Internal configuration populated by ConfigOption(s).
	config appConfig

	root    *node    // Internal command tree.
	plugins []Plugin // Registered plugins.
	globals []Flag   // global flags
}

// appConfig holds non-exported settings modified through ConfigOption.
type appConfig struct {
	debug        bool
	log          *log.Logger
	trace        bool
	panicHandler func(any)
}

// Command represents a runnable sub-command. Name and Aliases are Only
// Advisory; the actual registration path is determined by App.Command().
type Command struct {
	Name     string
	Aliases  []string
	Usage    string
	Short    string
	Long     string
	Category string

	Before func(*Context) error // Executed before Action.
	Action func(*Context) error // Required logic; must be non-nil.
	After  func(*Context) error // Executed after Action even if it errors.

	Flags *flag.FlagSet
}

// Plugin is the extension point for reusable behaviour such as
// middleware, extra commands or global flag injection
type Plugin interface {
	// Install is called once when the plugin is registered via App.Use.
	Sparkle(*App) error
}

// NotFoundHandler is invoked when no matching command is found.
type NotFoundHandler func(*Context, string) error

// ErrorHandler is invoked whenever Command.Action, Before, or After
// returns a non-nil error.
type ErrorHandler func(*Context, error) error

// node is the internal command tree nkde.
type node struct {
	cmd   *Command
	child map[string]*node
}

func (n *node) get(parts []string) (*node, []string) {
	cur := n
	for i, p := range parts {
		next, ok := cur.child[p]
		if !ok {
			return cur, parts[i:]
		}
		cur = next
	}
	return cur, nil
}

// --- internal helper ---
func isBuiltin(name string) bool {
	switch name {
	case "version", "help":
		return true
	}
	return false
}

func (a *App) debugf(format string, v ...any) {
	if !a.config.debug && !a.config.trace {
		return
	}
	a.config.log.Printf("[%s] %s", a.Name, fmt.Sprintf(format, v...))
}

// add inserts cmd into the tree at the given path.
func (a *App) add(path string, cmd *Command) (*App, error) {
	// root override
	if path == rootCommandPath {
		a.root.cmd = cmd
		return a, nil
	}

	parts := strings.Split(path, " ")
	cur, _ := a.root.get(parts[:len(parts)-1])
	name := parts[len(parts)-1]

	if _, ok := cur.child[name]; ok {
		if isBuiltin(name) {
			delete(cur.child, name)
		} else {
			return nil, fmt.Errorf("duplicate command: %s", path)
		}
	}

	cmd.Name = name
	cur.child[name] = &node{cmd: cmd, child: make(map[string]*node)}
	return a, nil
}

// New creates a fresh CLI application ready for configuration.
// The name should match your executable name (e.g. "git" or "docker").
// Can be configure with settings:
//
//	app := cli.New("app",
//	  cli.SetVersion("1.5"),
//	  cli.FluxDebug(true),
//	  // and more...
//	)
func New(name string, opts ...ConfigOption) *App {
	app := &App{
		Name: name,
		OnNotFound: func(ctx *Context, s string) error {
			fmt.Fprintf(ctx.App.Err, "command %s not found\n", s)
			return nil
		},
		OnError: func(ctx *Context, err error) error {
			fmt.Fprintln(ctx.App.Err, err)
			return err
		},
		Out:  os.Stdout,
		Err:  os.Stderr,
		root: &node{child: make(map[string]*node)},
		config: appConfig{
			debug: false,
			log:   log.New(os.Stderr, "[DEBUG] ", log.Ltime),
		},
	}

	for _, o := range opts {
		o(app)
	}

	app.Adopt(printAppVersion{})

	return app
}

// Command adds a new command to your CLI application.
// The path determines where the command lives in your command hierarchy.
// For example:
//
//	app.Command("server start", ...)  // Creates nested "server start" command
//	app.Command("status", ...)        // Creates top-level "status" command
//
// You can provide either an action function or configuration options:
//
//	app.Command("hello", func(c *cli.Context) error { ... })
//	app.Command("hello", cli.Action(...), cli.Short("Greets the user"))
func (a *App) Command(path string, fn func(*Context) error, opts ...CommandOption) (*App, error) {
	cmd := &Command{Name: path, Action: fn}

	for _, o := range opts {
		o(cmd)
	}

	return a.add(path, cmd)
}

// Use registers zero or more plugins
//
//	app.Use(&plugin1{}, &plugin2{}, ...)
func (a *App) Adopt(p ...Plugin) *App {
	for i, pl := range p {
		if pl == nil {
			a.debugf("plugin at index %d is nil", i)
			continue
		}

		if err := pl.Sparkle(a); err != nil {
			panic(err)
		}
	}
	return a
}

// --- execution helpers ---
func (a *App) execute(c *Command, args []string) (err error) {
	if c == nil {
		if len(args) == 0 && a.root.cmd != nil {
			c = a.root.cmd
		}
		return fmt.Errorf("no command defined: status Nil Command")
	}

	if c.Flags == nil {
		c.Flags = flag.NewFlagSet(c.Name, flag.ContinueOnError)
	}

	if c == a.root.cmd {
		args = append([]string{""}, args...)
	}

	fs := c.Flags

	for _, gf := range a.globals {
		gf.apply(fs)
	}

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	// validate required flags & ranges
	c.Flags.VisitAll(func(f *flag.Flag) {
		req, ok := f.Value.(interface{ Required() bool })
		if ok && req.Required() {
			if !isFlagPassed(c.Flags, f.Name) {
				err = fmt.Errorf("required flag --%s not provided", f.Name)
			}
		}

		v, ok := f.Value.(interface{ Validate() error })
		if ok {
			e := v.Validate()
			if e != nil {
				err = e
			}
		}
	})

	if c.Action == nil {
		return fmt.Errorf("no action defined for: %s", c.Name)
	}

	ctx := &Context{
		App:     a,
		Cmd:     c,
		RawArgs: args,
		Flags:   fs,
	}

	if c.Before != nil {
		if err = c.Before(ctx); err != nil {
			return err
		}
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

// internal recover wrapper
func (a *App) safeExecute(c *Command, args []string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if a.config.panicHandler != nil {
				a.config.panicHandler(r)
			} else {
				err = fmt.Errorf("panic: %v", r)
			}
		}
	}()

	return a.execute(c, args)
}

func (a *App) Parse(args []string) error {
	a.debugf("bug report: https://github.com/fyrna/cli/issues")

	if len(args) == 0 {
		a.debugf("no root command set yet")

		// 1) root command
		if a.root.cmd != nil {
			a.debugf("executing root command override")
			return a.safeExecute(a.root.cmd, []string{rootCommandPath})
		}

		// 2) help command
		h, ok := a.root.child["help"]
		if ok && h.cmd != nil {
			a.debugf("falling back to help command")
			return a.safeExecute(h.cmd, []string{"help"})
		}

		// 3) default
		a.debugf("showing default root help")
		return a.PrintRootHelp()
	}

	// Check if the first argument is a known command
	// and NOT a root command
	n, _ := a.root.get(args)
	if n.cmd != nil && n.cmd.Name != "" {
		return a.safeExecute(n.cmd, args)
	}

	// If we get here, it's either:
	// 1. A global flag
	// 2. An unknown command
	if a.root.cmd != nil && strings.HasPrefix(args[0], "-") {
		return a.safeExecute(a.root.cmd, args)
	}

	// Otherwise show command not found
	return a.OnNotFound(&Context{App: a}, args[0])
}

// Run executes the application with os.Args and handles errors
func (a *App) Run() {
	if err := a.Parse(os.Args[1:]); err != nil {
		ctx := &Context{App: a}
		if err2 := a.OnError(ctx, err); err2 != nil {
			a.config.log.Printf("OnError returned: %v", err2)
		}
		os.Exit(1)
	}
}
