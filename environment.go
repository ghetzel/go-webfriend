package webfriend

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/ghetzel/friendscript"
	"github.com/ghetzel/friendscript/commands/file"
	"github.com/ghetzel/friendscript/scripting"
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/commands/cookies"
	"github.com/ghetzel/go-webfriend/commands/core"
	"github.com/ghetzel/go-webfriend/commands/page"
)

var MaxReaderWait = time.Duration(5) * time.Second

type Environment struct {
	*friendscript.Environment
	Cookies *cookies.Commands
	Core    *core.Commands
	Page    *page.Commands
	File    *file.Commands
	browser *browser.Browser
	script  *scripting.Friendscript
	stack   []*scripting.Scope
}

func NewEnvironment(b *browser.Browser) *Environment {
	environment := &Environment{
		Environment: friendscript.NewEnvironment(),
		browser:     b,
		stack:       make([]*scripting.Scope, 0),
	}

	if b := environment.browser; b != nil {
		b.SetScope(environment)

		b.Tab().RegisterEventHandler(`Inspector.detached`, func(event *browser.Event) {
			b.Stop()
			os.Exit(0)
		})
	}

	// add in our custom modules and module overrides
	environment.Core = core.New(b)
	environment.Cookies = cookies.New(b)
	environment.Page = page.New(b)

	environment.RegisterModule(``, environment.Core)
	environment.RegisterModule(`cookies`, environment.Cookies)
	environment.RegisterModule(`page`, environment.Page)

	// add our custom REPL commands
	environment.RegisterCommandHandler(`help`, environment.handleReplHelp)

	// add command context handlers
	environment.RegisterContextHandler(func(ctx *scripting.Context, isCompleted bool) {
		if browser := environment.Browser(); browser != nil {
			params := map[string]interface{}{
				`command`: ctx.Label,
				`offset`:  ctx.AbsoluteStartOffset,
				`advance`: ctx.Length,
			}

			if id := environment.Scope().Get(`invocation`); !typeutil.IsZero(id) {
				params[`id`] = id
			}

			if isCompleted {
				params[`action`] = `finished`
				params[`took`] = (ctx.Took.Round(time.Microsecond) / time.Millisecond)
			} else {
				params[`action`] = `running`
			}

			if err := ctx.Error; err != nil {
				params[`error`] = err.Error()
			}

			browser.Tab().Emit(`Webfriend.scriptContextEvent`, params)

			if isCompleted && ctx.Error == nil {
				if delay := browser.Tab().AfterCommandDelay; delay > 0 {
					time.Sleep(delay)
				}
			}
		}
	})

	return environment
}

func (self *Environment) MustModule(name string) friendscript.Module {
	if module, ok := self.Module(name); ok {
		return module
	} else {
		panic(fmt.Sprintf("Invalid module %q", name))
	}
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
