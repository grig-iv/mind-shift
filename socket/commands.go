package socket

import "github.com/grig-iv/mind-shift/domain"

type Cmd interface{}

type QuitCmd struct{}

type KillClientCmd struct{}

type GoToTagCmd struct {
	Dir domain.Dir
}

type MoveToTagCmd struct {
	Dir domain.Dir
}

type GoToWinOrSpawn struct {
	Instance string
	Class    string
	SpanCmd  string
	SpanArgs []string
}

type UnknownCmd struct {
	Command string
}
