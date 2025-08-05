package cli

import "strings"

// Args represents the non-flag positional arguments of a command.
type Args []string

// Args returns the slice of Args
func (c *Context) Args() Args {
	return Args(c.Flags.Args())
}

// Get returns the i-th positional argument or empty string if
// the index is invalid.
func (a Args) Get(i int) string {
	if i < 0 || i >= len(a) {
		return ""
	}

	return a[i]
}

// String returns all positional arguments as a single space-separated string.
func (a Args) String() string {
	return strings.Join(a, " ")
}
