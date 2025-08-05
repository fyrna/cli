package cli

import (
	"flag"
	"strconv"
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

func (c *Context) GetString(name string) string {
	if c.Flags == nil {
		return ""
	}
	if val := c.Flags.Lookup(name); val != nil {
		return val.Value.String()
	}
	return ""
}

func (c *Context) GetBool(name string) bool {
	if c.Flags == nil {
		return false
	}
	if val := c.Flags.Lookup(name); val != nil {
		if b, err := strconv.ParseBool(val.Value.String()); err == nil {
			return b
		}
	}
	return false
}

func (c *Context) GetInt(name string) int {
	if c.Flags == nil {
		return 0
	}
	if val := c.Flags.Lookup(name); val != nil {
		if i, err := strconv.Atoi(val.Value.String()); err == nil {
			return i
		}
	}
	return 0
}

func (c *Context) GetFloat64(name string) float64 {
	v, _ := strconv.ParseFloat(c.Flags.Lookup(name).Value.String(), 64)
	return v
}

// func (c *Context) GetString(name string) string {
// 	return c.Flags.Lookup(name).Value.(flag.Getter).Get().(string)
// }
// func (c *Context) GetBool(name string) bool {
// 	return c.Flags.Lookup(name).Value.(flag.Getter).Get().(bool)
// }
// func (c *Context) GetInt(name string) int {
// 	return c.Flags.Lookup(name).Value.(flag.Getter).Get().(int)
// }
