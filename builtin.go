package cli

import "fmt"

// default help using app.Out as its output
func (a *App) PrintRootHelp() error {
	if a.Version != "" {
		fmt.Fprintf(a.Out, "%s - v%s\n", a.Name, a.Version)
	} else {
		fmt.Fprintf(a.Out, "%s\n", a.Name)
	}

	if a.Desc != "" {
		fmt.Fprintf(a.Out, "\n%s\n", a.Desc)
	}

	return nil
}

type BuiltinPlugin struct{}

func (p BuiltinPlugin) Sparkle(a *App) error {
	// Honour user-supplied "version" command.
	if _, ok := a.root.child["version"]; ok {
		return nil
	}

	a.Command("version", func(c *Context) error {
		if c.App.Version == "" {
			return fmt.Errorf("version not set")
		}
		_, err := fmt.Fprintln(c.App.Out, c.App.Version)
		return err
	})

	// builtin help flag
	a.Flags(Bool("help", "h").Help("show help."))

	return nil
}
