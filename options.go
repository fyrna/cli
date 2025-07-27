package cli

import (
	"flag"
)

// :: config
type ConfigOption func(*App)

func SetVersion(v string) ConfigOption {
	return func(a *App) { a.Version = string(v) }
}

func SetDesc(d string) ConfigOption {
	return func(a *App) { a.Desc = d }
}

func SetDebug(e bool) ConfigOption {
	return func(a *App) { a.config.debug = e }
}

// :: command
type CommandOption func(*Command)
type fun func(*Context) error

// what defined command should do?
func Action(fn fun) CommandOption {
	return func(c *Command) { c.Action = fn }
}

// execute before action
func Before(fn fun) CommandOption {
	return func(c *Command) { c.Before = fn }
}

// execute after action
func After(fn fun) CommandOption {
	return func(c *Command) { c.After = fn }
}

func Short(s string) CommandOption {
	return func(c *Command) { c.Short = s }
}

func Long(s string) CommandOption {
	return func(c *Command) { c.Long = s }
}

func Alias(a ...string) CommandOption {
	return func(c *Command) { c.Aliases = a }
}

func Usage(u string) CommandOption {
	return func(c *Command) { c.Usage = u }
}

func Category(cat string) CommandOption {
	return func(c *Command) { c.Category = cat }
}

func Flags(fs *flag.FlagSet) CommandOption {
	return func(c *Command) { c.Flags = fs }
}
