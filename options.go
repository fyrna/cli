package cli

import (
	"flag"
	"io"
	"log"
)

// :: config
type ConfigOption func(*App)

// app config

// set app version
func SetVersion(v string) ConfigOption {
	return func(a *App) { a.Version = string(v) }
}

// set app description
func SetDesc(d string) ConfigOption {
	return func(a *App) { a.Desc = d }
}

// internal config

// set debug
func FluxDebug(on bool) ConfigOption {
	return func(a *App) { a.config.debug = on }
}

// custom output for debug option
// using's std log.SetOutput()
func FluxDebugOutput(w io.Writer) ConfigOption {
	return func(a *App) {
		a.config.log.SetOutput(w)
	}
}

// custom logger
// should compatible and/or just use stdlib's logger
func FluxLogger(l *log.Logger) ConfigOption {
	return func(a *App) { a.config.log = l }
}

// tracing
func FluxTrace(on bool) ConfigOption {
	return func(a *App) { a.config.trace = on }
}

// set panic handler
func FluxPanicHandler(fn func(any)) ConfigOption {
	return func(a *App) { a.config.panicHandler = fn }
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

// help me add comment to every pub function :(
