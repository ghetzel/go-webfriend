package webfriend

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/ghetzel/friendscript"
	"github.com/ghetzel/friendscript/scripting"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/commands/assert"
	"github.com/ghetzel/go-webfriend/commands/cookies"
	"github.com/ghetzel/go-webfriend/commands/core"
	"github.com/ghetzel/go-webfriend/commands/page"
)

var MaxReaderWait = time.Duration(5) * time.Second

type Environment struct {
	*friendscript.Environment
	Assert  *assert.Commands
	Cookies *cookies.Commands
	Core    *core.Commands
	Page    *page.Commands
	browser *browser.Browser
	script  *scripting.Friendscript
	stack   []*scripting.Scope
}

func NewEnvironment(browser *browser.Browser) *Environment {
	environment := &Environment{
		Environment: friendscript.NewEnvironment(),
		browser:     browser,
		stack:       make([]*scripting.Scope, 0),
	}

	if environment.browser != nil {
		environment.browser.SetScope(environment)
	}

	// add in our custom modules and module overrides
	environment.Core = core.New(browser, environment)
	environment.Assert = assert.New(browser, environment)
	environment.Cookies = cookies.New(browser, environment)
	environment.Page = page.New(browser, environment)

	environment.RegisterModule(``, environment.Core)
	environment.RegisterModule(`assert`, environment.Assert)
	environment.RegisterModule(`cookies`, environment.Cookies)
	environment.RegisterModule(`page`, environment.Page)

	// add our custom REPL commands
	environment.RegisterCommandHandler(`help`, environment.handleReplHelp)

	return environment
}

func (self *Environment) Browser() *browser.Browser {
	return self.browser
}

func (self *Environment) handleReplHelp(ctx *friendscript.InteractiveContext, environment *friendscript.Environment) ([]string, error) {
	// This is a big colorized Unicode WebFriend logo for terminal output.

	// [ .. ] hi blue w/ blue bg
	// { .. } hi green w/ blue bg
	// ( .. ) hi black w/ blue bg
	// < .. > hi green w/ default bg

	lines := []string{}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("[Web]<friend> v%s - %q", Version, Slogan))
	lines = append(lines, "")
	lines = append(lines, "                       [░░▓▓▓▓▓▓▓▓▓▓▓░░]")
	lines = append(lines, "                   [░▓▓█████████████████▓▓░]")
	lines = append(lines, "               {░████▙}[██████████████████████▓▓░░]")
	lines = append(lines, "            {░███████▛}[██]{▟██▙}[███████████████]{▟█████░}")
	lines = append(lines, "          {░██████████}[█]{▟██▛}[███████████████]{▟█▛██████░}")
	lines = append(lines, "        {░████████████████}[██████████████████]{▟████████░}")
	lines = append(lines, "       {░███████████████▛}[████████████████]{▟████████████░}")
	lines = append(lines, "      {░█████████████▛}[█████████████████]{▟██████▛}[█]{▜█▛}[█]{███░}")
	lines = append(lines, "     {░█████████████▛}[█](▟██▙)[█████████](▟██▙){▜██▛}[█████]{▜▛}[██]{████░}")
	lines = append(lines, "     {░█}[█]{▜████▛}[██]{▜█▙}[██](████)[█████████](████)[█████████████]{████░}")
	lines = append(lines, "     [░███]{▜████}[███████](▜██▛)[█████████](▜██▛){▟█████████████████░}")
	lines = append(lines, "     [░█████]{▜█▙}[████████████████████]{▟████████████████████░}")
	lines = append(lines, "     [░██████]{▜█▙}[██████████████████]{█████████████████████▓░}")
	lines = append(lines, "      [░▓▓█]{▟██████▙}[████████████████]{▜███████████████▛}[██▓░]")
	lines = append(lines, "       {░███████████▙}[███████████████████]{▟██████████▛}[█▓░]")
	lines = append(lines, "        {░███████████}[█████](▜████████▛)[███]{██████████▛}[█▓▓░]")
	lines = append(lines, "          {░███████▛}[███████](▜██████▛)[████]{█████████▛}[█░]")
	lines = append(lines, "            {░████▛}[███████████████████]{▟███████▛}[██░]")
	lines = append(lines, "               [░▓▓███████████████████]{███████}[█░]")
	lines = append(lines, "                   [░▓▓███████████████]{▜█▓}[▓░]")
	lines = append(lines, "                       [░░▓▓▓▓▓▓▓▓▓▓▓▓░░]")

	output := ``

	for i, line := range lines {
		outline := ``
		state := 0

		for _, c := range line {
			switch c {
			case '[':
				state = 1
			case '{':
				state = 2
			case '(':
				state = 3
			case '<':
				state = 4
			case ']', '}', ')', '>':
				switch state {
				case 1:
					output += color.New(color.FgHiBlue).Sprint(outline)
				case 2:
					output += color.New(color.FgHiGreen, color.BgHiBlue).Sprint(outline)
				case 3:
					output += color.New(color.FgHiBlack, color.BgHiBlue).Sprint(outline)
				case 4:
					output += color.New(color.FgHiGreen).Sprint(outline)
				}

				state = 0
				outline = ``
			default:
				if state > 0 {
					outline += string(c)
				} else {
					output += string(c)
				}
			}
		}

		lines[i] = output
	}

	return lines, nil
}
