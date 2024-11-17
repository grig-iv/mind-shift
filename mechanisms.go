package main

import "github.com/grig-iv/mind-shift/x"

func (wm *windowManager) enableFullscreen(client *client) {
	client.isFullscreen = true
	x.ChangeBorderWidth(client.window, 0)
	x.Raise(client.window)
	wm.view(wm.currTag)
}

func (wm *windowManager) disableFullscreen(client *client) {
	client.isFullscreen = false
	x.ChangeBorderWidth(client.window, borderWidth)
	wm.view(wm.currTag)
}
