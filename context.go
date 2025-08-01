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
	Flags   *flag.FlagSet
}

// Exec re-parses the supplied path and arguments as if they came from the real
// command line. This allows commands to programmatically invoke another commands.
//
//	 app.Command("greet", func(c *cli.Context) error {
//		  fmt.Println("hello, im greet command")
//		  return nil
//		})
//
//	 app.Command("hello", func(c *cli.Context) error {
//		  fmt.Println("original hello")
//		  c.Exec("greet")
//		  return nil
//	 })
func (c *Context) Exec(path string, args ...string) error {
	parts := append(strings.Split(path, " "), args...)
	return c.App.Parse(parts)
}
