// fyrna/cli is a tiny, flexible command-line micro-framework.
package cli

import (
	"fmt"
)

// PrintRootHelp prints a concise overview of the application
// (name, version, and description) to App.Out.
func (a *App) PrintRootHelp() error {
	if a.Version != "" {
		fmt.Fprintf(a.Out, "%s %s\n", a.Name, a.Version)
	} else {
		fmt.Fprintf(a.Out, "%s\n", a.Name)
	}

	if a.Desc != "" {
		fmt.Fprintf(a.Out, "\n%s\n", a.Desc)
	}

	return nil
}

// printAppVersion is a built-in plugin that adds a "version"
// command unless the user has already defined one.
type printAppVersion struct{}

func (pav printAppVersion) Install(a *App) error {
	// Honour user-supplied "version" command.
	if _, ok := a.root.subs["version"]; ok {
		return nil
	}

	a.Command("version", func(c *Context) error {
		if c.App.Version == "" {
			return fmt.Errorf("version not set")
		}
		fmt.Fprintln(c.App.Out, c.App.Version)
		return nil
	})

	return nil
}
