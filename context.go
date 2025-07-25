package cli

import "flag"

// Context carries data through the call chain.
type Context struct {
	App     *App
	Cmd     *Command
	RawArgs []string
	Store   map[string]any
	Flags   *flag.FlagSet
}
