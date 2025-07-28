package cli

import (
	"fmt"
)

// minimal help
func (a *App) ShowRootHelp() error {
	if a.Version != "" {
		fmt.Fprintf(a.Out, "%s - %s\n", a.Name, a.Version)
	} else {
		fmt.Fprintf(a.Out, "%s\n", a.Name)
	}

	if a.Desc != "" {
		fmt.Fprintf(a.Out, "\t%s\n", a.Desc)
	}

	return nil
}

// print version
type printAppVersion struct{}

// youre my pav <3
func (pav printAppVersion) Install(a *App) error {
	if _, ok := a.root.subs["version"]; ok {
		return nil
	}

	a.Command("version", func(c *Context) error {
		if c.App.Version != "" {
			fmt.Fprintf(c.App.Out, "version %s", c.App.Version)
			return nil
		}

		return fmt.Errorf("please set your app version")
	})

	return nil
}
