// fyrna/cli is a tiny, flexible command-line micro-framework.
package cli

import (
	"flag"
	"strings"
)

// Context carries request-scoped data accross Before, Action, and After hooks.
type Context struct {
	App     *App     // Reference to the CLI application.
	Cmd     *Command // The command currently being executed.
	RawArgs []string // Unprocessed arguments (including name).
	Store   map[string]any
	Flags   *flag.FlagSet
}

// Exec re-parses the supplied path and arguments as if they came from the real
// command line. This allows commands to programmatically invoke another commands.
//
//	app.Command("greet", ...)
//
//	app.Command("hello", func(c *cli.Context) error {
//	    c.Exec("greet")
//	})
func (c *Context) Exec(path string, args ...string) error {
	parts := append(strings.Split(path, " "), args...)
	return c.App.Parse(parts)
}
