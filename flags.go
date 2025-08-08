package cli

import (
	"flag"
	"fmt"
)

// Flag represents a command line flag that can be attached
// to commands. Use the spesific flag types like String(), Bool(),
// or Int() to create flags with different value types.
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

func (a *App) Flags(ff ...Flag) *App {
	a.globals = append(a.globals, ff...)
	return a
}

func Flags(ff ...Flag) CommandOption {
	return func(cmd *Command) {
		for _, f := range ff {
			f.apply(flagSet(&cmd.Flags))
		}
	}
}

// FlagInfo exposes the minimal read-only view of a flag.
type FlagInfo interface {
	GetName() string         // long name, e.g. "config"
	GetUsage() string        // help text
	GetShort() []string      // short name, e.g. "c"
	GetDefaultValue() string // default value as string
	HasShort() bool          // true if has short form
	IsBool() bool            // true if boolean flag
}

// i have no idea how to do this actually, so, here you go.
//
// READ-ONLY HELPER
func (f *stringFlag) GetName() string {
	return f.name
}
func (f *stringFlag) GetUsage() string {
	return f.usage
}
func (f *stringFlag) GetShort() []string {
	if len(f.short) > 0 {
		return f.short
	}
	return nil
}
func (f *stringFlag) GetDefaultValue() string {
	return f.def
}
func (f *stringFlag) IsBool() bool {
	return false
}
func (f *stringFlag) HasShort() bool {
	return len(f.short) > 0
}
func (f *boolFlag) GetName() string {
	return f.name
}
func (f *boolFlag) GetUsage() string {
	return f.usage
}
func (f *boolFlag) GetShort() []string {
	if len(f.short) > 0 {
		return f.short
	}
	return nil
}
func (f *boolFlag) GetDefaultValue() string {
	return fmt.Sprintf("%t", f.def)
}
func (f *boolFlag) HasShort() bool {
	return len(f.short) > 0
}
func (f *boolFlag) IsBool() bool {
	return true
}
