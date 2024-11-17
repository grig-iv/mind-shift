package socket

import "github.com/grig-iv/mind-shift/domain"

type parser func(args []string) Cmd

var parsers = map[string]parser{
	"quit":        simpleParser(QuitCmd{}),
	"kill-client": simpleParser(KillClientCmd{}),
	"full-screen": simpleParser(FullScreenCmd{}),

	"go-to-tag":          goToTagParser,
	"move-to-tag":        moveToTagParser,
	"go-to-win-or-spawn": goToWinOrSpawnParser,
}

func simpleParser(cmd Cmd) parser {
	return func(_ []string) Cmd {
		return cmd
	}
}

func goToTagParser(args []string) Cmd {
	if len(args) != 1 {
		return InvalidArgsCmd{}
	}

	switch args[0] {
	case "-n":
		return GoToTagCmd{Dir: domain.Next}
	case "-p":
		return GoToTagCmd{Dir: domain.Prev}
	default:
		return InvalidArgsCmd{}
	}
}

func moveToTagParser(args []string) Cmd {
	if len(args) != 1 {
		return InvalidArgsCmd{}
	}

	switch args[0] {
	case "-n":
		return MoveToTagCmd{Dir: domain.Next}
	case "-p":
		return MoveToTagCmd{Dir: domain.Prev}
	default:
		return InvalidArgsCmd{}
	}
}

func goToWinOrSpawnParser(args []string) Cmd {
	if len(args) < 2 {
		return InvalidArgsCmd{}
	}

	return GoToWinOrSpawnCmd{
		Class:    args[0],
		SpanCmd:  args[1],
		SpanArgs: args[2:],
	}
}
