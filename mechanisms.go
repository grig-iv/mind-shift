package main

import (
	"log"

	"github.com/grig-iv/mind-shift/x"
)

func (wm *windowManager) enableFullscreen(client *client) {
	log.Println("Enabling fullscreen. client:", client.window)
	client.isFullscreen = true
	x.ChangeBorderWidth(client.window, 0)
	x.Raise(client.window)
	wm.view(wm.currTag)
}

func (wm *windowManager) disableFullscreen(client *client) {
	log.Println("Disabling fullscreen. client:", client.window)
	client.isFullscreen = false
	x.ChangeBorderWidth(client.window, borderWidth)
	wm.view(wm.currTag)
}
