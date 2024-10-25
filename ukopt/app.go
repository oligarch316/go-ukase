package ukopt

import "github.com/oligarch316/ukase"

// =============================================================================
// General
// =============================================================================

var _ ukase.Option = App(nil)

type App func(*ukase.Config)

func (o App) UkaseApplyApp(c *ukase.Config) { o(c) }

// =============================================================================
// Specific
// =============================================================================

func AppHelpCommand(name string) App      { return func(c *ukase.Config) { c.HelpCommand = name } }
func AppInputProgram(name string) App     { return func(c *ukase.Config) { c.InputProgram = name } }
func AppInputArguments(args []string) App { return func(c *ukase.Config) { c.InputArguments = args } }
