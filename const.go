package cli

const (
	rootCommandPath = "" // internal sentinel for root override
)

// Debug/trace messages.
const (
	debugReport           = "report bug: https://github.com/fyrna/cli/issues"
	debugNoRootCommand    = "no root command set yet"
	debugNoHelpCommand    = "no help command provided"
	debugUsingDefaultHelp = "using default help command"
	debugArgsParsed       = "onii-chan, args parsed: %s"
)

// User-facing error messages.
const (
	errNilCommand       = "nil command"
	errNilPlugin        = "nil plugin"
	errNoAction         = "onii-chan, no action defined for %s"
	errCommandNotFound  = "command %s not found\n"
	errDuplicateCommand = "duplicate command: %s"
)
