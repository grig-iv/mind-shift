package main

import (
	"log"
	"os/exec"

	"github.com/grig-iv/mind-shift/x"
	"github.com/jezek/xgb/xproto"
)

func (wm *windowManager) enableFullscreen(client *client) {
	log.Println("[wm.enableFullscreen]", client.window)
	x.ChangeProperty32(
		client.window,
		x.AtomOrNone(x.NetWMState),
		xproto.AtomAtom,
		uint(x.AtomOrNone(x.NetWMFullscreen)),
	)
	client.isFullscreen = true
	x.ChangeBorderWidth(client.window, 0)
	x.Raise(client.window)
	wm.view(wm.currTag)
}

func (wm *windowManager) disableFullscreen(client *client) {
	log.Println("[wm.disableFullscreen]", client.window)
	x.ChangeProperty32(
		client.window,
		x.AtomOrNone(x.NetWMState),
		xproto.AtomAtom,
		0,
	)
	client.isFullscreen = false
	x.ChangeBorderWidth(client.window, borderWidth)
	wm.view(wm.currTag)
}

func (wm *windowManager) findClientByClass(targetClass string) (*client, bool) {
	for _, c := range wm.clients {
		_, clientClass := x.InstanceAndClass(c.window)
		if targetClass == clientClass {
			return c, true
		}
	}

	return nil, false
}

func (wm *windowManager) spawnIfNotExist(targetClass string, command string, args ...string) {
	_, found := wm.findClientByClass(targetClass)
	if !found {
		exec.Command(command, args...).Start()
	}
}
