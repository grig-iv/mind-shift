package main

import (
	"log"

	"github.com/grig-iv/mind-shift/domain"
	"github.com/grig-iv/mind-shift/socket"
)

func (wm *windowManager) eval(cmd socket.Cmd) {
	switch cmd := cmd.(type) {

	case socket.GoToTagCmd:
		if cmd.Dir == domain.Prev {
			wm.goToPrevTag()
		} else {
			wm.goToNextTag()
		}

	case socket.GoToWinOrSpawnCmd:
		wm.gotoWindowOrSpawn(cmd.Class, cmd.SpanCmd, cmd.SpanArgs...)

	case socket.MoveToTagCmd:
		if cmd.Dir == domain.Next {
			wm.moveToNextTag()
		} else {
			wm.moveToPrevTag()
		}

	case socket.KillClientCmd:
		wm.killClient()

	case socket.FullScreenCmd:
		if wm.focusedClient.isFullscreen {
			wm.disableFullscreen(wm.focusedClient)
		} else {
			wm.enableFullscreen(wm.focusedClient)
		}

	case socket.QuitCmd:
		wm.quit()

	default:
		log.Println("Unknown command:", cmd)
	}
}
