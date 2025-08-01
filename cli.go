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

// App is the root CLI application.
type App struct {
	Name    string
	Version string
	Desc    string

	// behaviour hooks
	OnNotFound NotFoundHandler
	OnError    ErrorHandler

	// I/O streams used by the framework and user handlers.
	// Default are os.Stdout and os.Stderr respectively.
	Out io.Writer
	Err io.Writer

	// Internal configuration populated by ConfigOption(s).
	config appConfig

	root    *node    // Internal command tree.
	plugins []Plugin // Registered plugins.
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
	Install(*App) error
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
func (a *App) add(path string, cmd *Command) *App {
	// root override
	if path == rootCommandPath {
		a.root.cmd = cmd
		return a
	}

	parts := strings.Split(path, " ")
	cur, _ := a.root.get(parts[:len(parts)-1])
	name := parts[len(parts)-1]

	if _, ok := cur.child[name]; ok {
		if isBuiltin(name) {
			delete(cur.child, name)
		} else {
			panic(fmt.Sprintf(errDuplicateCommand, path))
		}
	}

	cmd.Name = name
	cur.child[name] = &node{cmd: cmd, child: make(map[string]*node)}
	return a
}

// Constructor
func New(name string, opts ...ConfigOption) *App {
	app := &App{
		Name: name,
		OnNotFound: func(ctx *Context, s string) error {
			fmt.Fprintf(ctx.App.Err, errCommandNotFound, s)
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

	app.Use(printAppVersion{})

	return app
}

// Command registers a new sub-command at the given path.
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

// Use registers zero or more plugins
//
//	app.Use(plugin1(), plugin2(), ...)
func (a *App) Use(p ...Plugin) *App {
	for i, pl := range p {
		if pl == nil {
			a.debugf("plugin at index %d is nil", i)
			continue
		}

		if err := pl.Install(a); err != nil {
			panic(err)
		}
	}
	return a
}

// --- execution helpers ---
func (a *App) execute(c *Command, args []string) (err error) {
	if c == nil {
		return fmt.Errorf(errNilCommand)
	}

	fs := c.Flags
	if fs == nil {
		fs = flag.NewFlagSet(c.Name, flag.ContinueOnError)
	}

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	if c.Action == nil {
		return fmt.Errorf(errNoAction, c.Name)
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
	a.debugf("%s", debugReport)

	if len(args) == 0 {
		a.debugf("%s", debugNoRootCommand)

		// 1) root command
		if a.root.cmd != nil {
			a.debugf("executing root command override")
			return a.execute(a.root.cmd, []string{rootCommandPath})
		}

		// 2) help command
		h, ok := a.root.child["help"]
		if ok && h.cmd != nil {
			a.debugf("falling back to help command")
			return a.execute(h.cmd, []string{"help"})
		}

		// 3) default
		a.debugf("showing default root help")
		return a.PrintRootHelp()
	}

	n, rest := a.root.get(args)

	if n.cmd != nil {
		a.debugf(debugArgsParsed, args)
		return a.safeExecute(n.cmd, args)
	}

	ctx := &Context{App: a, RawArgs: rest}
	name := args[0]
	if len(rest) > 0 {
		name = rest[0]
	}

	return a.OnNotFound(ctx, name)
}

func (a *App) Run() {
	if err := a.Parse(os.Args[1:]); err != nil {
		ctx := &Context{App: a}
		if err2 := a.OnError(ctx, err); err2 != nil {
			a.config.log.Printf("OnError returned: %v", err2)
		}
		os.Exit(1)
	}
}
