package socket

import "github.com/grig-iv/mind-shift/domain"

type Cmd interface{}

type QuitCmd struct{}

type KillClientCmd struct{}

type FullScreenCmd struct{}

type GoToTagCmd struct {
	Dir domain.Dir
}

type MoveToTagCmd struct {
	Dir domain.Dir
}

type GoToWinOrSpawnCmd struct {
	Class    string
	SpanCmd  string
	SpanArgs []string
}

type UnknownCmd struct {
	Command string
}

type InvalidArgsCmd struct {
	Command string
	Args    []string
}
