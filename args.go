package cli

type Args []string

func (c *Context) Args() Args {
	return Args(c.Flags.Args())
}

func (a Args) Get(i int) string {
	if i < 0 || i >= len(a) {
		return ""
	}

	return a[i]
}
