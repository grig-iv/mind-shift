package main

import (
	"log"

	"github.com/grig-iv/mind-shift/domain"
	"github.com/grig-iv/mind-shift/socket"
)

func (wm *windowManager) eval(cmd socket.Cmd) {
	switch cmd := cmd.(type) {
	case socket.QuitCmd:
		wm.quit()
	case socket.KillClientCmd:
		wm.killClient()
	case socket.GoToTagCmd:
		if cmd.Dir == domain.Prev {
			wm.gotoNextTag()
		} else {
			wm.gotoPrevTag()
		}
	case socket.MoveToTagCmd:
		if cmd.Dir == domain.Next {
			wm.moveToNextTag()
		} else {
			wm.moveToPrevTag()
		}
	case socket.GoToWinOrSpawn:
		wm.gotoWindowOrCreate(cmd.Class, cmd.SpanCmd, cmd.SpanArgs...)
	default:
		log.Println("Unknown command:", cmd)
	}
}
