package main

import (
	"log"

	"github.com/grig-iv/mind-shift/commands"
)

func (wm *windowManager) eval(cmd commands.Cmd) {
	switch cmd := cmd.(type) {
	case commands.QuitCmd:
		wm.quit()
	case commands.KillClientCmd:
		wm.killClient()
	case commands.GoToTagCmd:
		if cmd.Dir == commands.Next {
			wm.gotoNextTag()
		} else {
			wm.gotoPrevTag()
		}
	case commands.MoveToTagCmd:
		if cmd.Dir == commands.Next {
			wm.moveToNextTag()
		} else {
			wm.moveToPrevTag()
		}
	case commands.GoToWinOrSpawn:
		wm.gotoWindowOrCreate(cmd.Class, cmd.SpanCmd, cmd.SpanArgs...)
	default:
		log.Println("Unknown command:", cmd)
	}
}
