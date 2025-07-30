package cli

import "strings"

type Args []string

// returns slice of Args
func (c *Context) Args() Args {
	return Args(c.Flags.Args())
}

// get spesific arg by index
func (a Args) Get(i int) string {
	if i < 0 || i >= len(a) {
		return ""
	}

	return a[i]
}

// returns all args
func (a Args) All() string {
	return strings.Join(a, " ")
}
