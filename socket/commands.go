package socket

type Cmd interface{}

type QuitCmd struct{}

type KillClientCmd struct{}

type GoToTagCmd struct {
	Dir Dir
}

type MoveToTagCmd struct {
	Dir Dir
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

type Dir byte

const (
	Prev Dir = iota
	Next
)
