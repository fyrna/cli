package cli

import (
	"flag"
	"strings"
)

// Context carries data through the call chain.
type Context struct {
	App     *App
	Cmd     *Command
	RawArgs []string
	Store   map[string]any
	Flags   *flag.FlagSet
}

func (c *Context) Exec(path string, args ...string) error {
	parts := append(strings.Split(path, " "), args...)
	return c.App.Parse(parts)
}
