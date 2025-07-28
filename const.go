// under MIT License, see LICENSE file
package cli

const rootCommandName = ""

const debugPrefix = "[cli:debug] "

// debug message
const (
	debugReport = debugPrefix + "report bug: https://github.com/fyrna/cli/issues"

	debugNoRootCommand    = debugPrefix + "no root command set yet"
	debugNoHelpCommand    = debugPrefix + "no help command provided"
	debugUsingDefaultHelp = debugPrefix + "using default help command"

	debugArgsParsed = debugPrefix + "onii-chan, args parsed: %s"
)

// err message
const (
	errCommandNotFound  = "command %s not found\n"
	errDuplicateCommand = "duplicate command: "
)
