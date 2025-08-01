package cli

import (
	"io"
	"log"
)

type ConfigOption func(*App)

// app config

// set app version
func SetVersion(v string) ConfigOption {
	return func(a *App) { a.Version = v }
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
// anything support io.Writer
func FluxDebugOutput(w io.Writer) ConfigOption {
	return func(a *App) {
		a.config.log.SetOutput(w)
	}
}

// custom logger
func FluxLogger(l *log.Logger) ConfigOption {
	return func(a *App) { a.config.log = l }
}

// set panic handler
func FluxPanicHandler(fn func(any)) ConfigOption {
	return func(a *App) { a.config.panicHandler = fn }
}

// config for command
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

// short description for defined command
func Short(s string) CommandOption {
	return func(c *Command) { c.Short = s }
}

// long description about defined command
func Long(s string) CommandOption {
	return func(c *Command) { c.Long = s }
}

// set aliases for command
func Alias(a ...string) CommandOption {
	return func(c *Command) { c.Aliases = a }
}

// information about command's usage
func Usage(u string) CommandOption {
	return func(c *Command) { c.Usage = u }
}

// categorizing command
func Category(cat string) CommandOption {
	return func(c *Command) { c.Category = cat }
}
