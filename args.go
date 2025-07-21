package cli

func (c *Context) Args() Args {
	return Args(c.Flags.Args())
}

type Args []string

func (a Args) Get(i int) string {
	if i < 0 || i >= len(a) {
		return ""
	}

	return a[i]
}
