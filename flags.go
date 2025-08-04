package cli

import (
	"flag"
	"fmt"
)

// Flag is implemented by the concrete flag builders below.
type Flag interface {
	apply(*flag.FlagSet)
}

func flagSet(fs **flag.FlagSet) *flag.FlagSet {
	if *fs == nil {
		*fs = flag.NewFlagSet("", flag.ContinueOnError)
	}
	return *fs
}

func isFlagPassed(fs *flag.FlagSet, name string) bool {
	return fs.Lookup(name).DefValue != fs.Lookup(name).Value.String()
}

// --- string ---
type stringFlag struct {
	name, usage string
	def         string
	short       []string
	required    bool
}

func String(name string, short ...string) *stringFlag {
	return &stringFlag{name: name, short: short}
}

func (f *stringFlag) Default(v string) *stringFlag {
	f.def = v
	return f
}

func (f *stringFlag) Help(h string) *stringFlag {
	f.usage = h
	return f
}

func (f *stringFlag) Required() *stringFlag {
	f.required = true
	return f
}

func (f *stringFlag) Validate() error {
	if f.required && f.def == "" {
		return fmt.Errorf("flag --%s is required", f.name)
	}
	return nil
}

func (f *stringFlag) apply(fs *flag.FlagSet) {
	if fs.Lookup(f.name) != nil {
		return // flag already exists
	}
	fs.StringVar(&f.def, f.name, f.def, f.usage)
	for _, s := range f.short {
		if fs.Lookup(s) == nil {
			fs.StringVar(&f.def, s, f.def, f.usage)
		}
	}
}

// --- bool ---
type boolFlag struct {
	name, usage   string
	short         []string
	def, required bool
}

func Bool(name string, short ...string) *boolFlag {
	return &boolFlag{name: name, short: short}
}

func (f *boolFlag) Help(h string) *boolFlag {
	f.usage = h
	return f
}

func (f *boolFlag) Required() *boolFlag {
	f.required = true
	return f
}

func (f *boolFlag) Validate() error {
	if f.required && !f.def {
		return fmt.Errorf("flag --%s is required", f.name)
	}
	return nil
}

func (f *boolFlag) apply(fs *flag.FlagSet) {
	if fs.Lookup(f.name) != nil {
		return // flag already exists
	}
	fs.BoolVar(&f.def, f.name, f.def, f.usage)
	for _, s := range f.short {
		if fs.Lookup(s) == nil {
			fs.BoolVar(&f.def, s, f.def, f.usage)
		}
	}
}

// --- int ---
type intFlag struct {
	name, usage string
	short       []string
	def         int
	min, max    int
}

func Int(name string) *intFlag {
	return &intFlag{name: name}
}

func (f *intFlag) Default(v int) *intFlag {
	f.def = v
	return f
}

func (f *intFlag) Range(min, max int) *intFlag {
	f.min, f.max = min, max
	return f
}

func (f *intFlag) Help(h string) *intFlag {
	f.usage = h
	return f
}

func (f *intFlag) Validate() error {
	if f.def < f.min || f.def > f.max {
		return fmt.Errorf("flag -%s value %d out of range [%d,%d]", f.name, f.def, f.min, f.max)
	}
	return nil
}

func (f *intFlag) apply(fs *flag.FlagSet) {
	if fs.Lookup(f.name) != nil {
		return
	}
	fs.IntVar(&f.def, f.name, f.def, f.usage)
	for _, s := range f.short {
		if fs.Lookup(s) == nil {
			fs.IntVar(&f.def, s, f.def, f.usage)
		}
	}
}

// GlobalFlags returns a CommandOption that adds flags to the root command.
func (a *App) Flags(ff ...Flag) *App {
	a.globals = append(a.globals, ff...)
	return a
}

// Flags returns a CommandOption that adds flags to the specific command.
func Flags(ff ...Flag) CommandOption {
	return func(cmd *Command) {
		for _, f := range ff {
			f.apply(flagSet(&cmd.Flags))
		}
	}
}
