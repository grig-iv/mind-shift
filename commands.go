package main

import (
	"log"
	"os/exec"
	"time"

	"github.com/jezek/xgb/xproto"
)

type direction byte

const (
	Forward  direction = iota
	Backward direction = iota
)

func (wm *windowManager) quit() {
	wm.isRunning = false
}

func (wm *windowManager) killClient() {
	if wm.focusedClient == nil {
		return
	}

	xproto.GrabServer(wm.x.conn)
	xproto.SetCloseDownMode(wm.x.conn, xproto.CloseDownDestroyAll)
	xproto.KillClient(wm.x.conn, uint32(wm.focusedClient.window))
	xproto.UngrabServer(wm.x.conn)
}

func (wm *windowManager) getPrevTag() tag {
	for i := range wm.tags {
		i = len(wm.tags) - i - 1
		if wm.tags[i] == wm.currTag {
			return wm.tags[(i-1+len(wm.tags))%len(wm.tags)]
		}
	}
	return tag{}
}

func (wm *windowManager) getNextTag() tag {
	for i := range wm.tags {
		if wm.tags[i] == wm.currTag {
			return wm.tags[(i+1)%len(wm.tags)]
		}
	}
	return tag{}
}

func (wm *windowManager) gotoPrevTag() {
	wm.view(wm.getPrevTag())
}

func (wm *windowManager) gotoNextTag() {
	wm.view(wm.getNextTag())
}

func (wm *windowManager) view(tag tag) {
	log.Println("[wm.view] tag number:", tag.index()+1)

	wm.currTag = tag

	tagClients := make([]*client, 0)
	for _, c := range wm.clients {
		if c.hasTag(wm.currTag.id) {
			tagClients = append(tagClients, c)
		} else {
			xproto.ConfigureWindow(
				c.conn,
				c.window,
				uint16(xproto.ConfigWindowX),
				[]uint32{uint32(c.geom.width * -2)},
			)
		}
	}

	screenGeom := geometry{
		x:      0,
		y:      0,
		width:  int(wm.x.screen.WidthInPixels),
		height: int(wm.x.screen.HeightInPixels),
	}

	screenGeom = wm.bar.adjustScreenGeometry(screenGeom)

	geoms := wm.currTag.currLaout.arrange(screenGeom, len(tagClients))
	for i, c := range tagClients {
		c.geom = geoms[i]
		wm.x.changeGeometry(c.window, c.geom)
	}

	if (wm.focusedClient == nil || wm.focusedClient.hasTag(tag.id) == false) && len(tagClients) > 0 {
		wm.focus(tagClients[0])
	}
}

func (wm *windowManager) moveToPrevTag() {
	wm.moveToTag(wm.getPrevTag())
}

func (wm *windowManager) moveToNextTag() {
	wm.moveToTag(wm.getNextTag())
}

func (wm *windowManager) moveToTag(tag tag) {
	if wm.focusedClient == nil || wm.currTag.id == tag.id {
		return
	}

	wm.focusedClient.replaceTag(wm.currTag.id, tag.id)

	for i, c := range wm.clients {
		if c == wm.focusedClient {
			wm.clients = append(wm.clients[:i], wm.clients[i+1:]...)
			wm.clients = append(wm.clients, c)
			break
		}
	}

	wm.view(tag)
}

func (wm *windowManager) gotoWindow(targetClass string) bool {
	for _, c := range wm.clients {
		_, clientClass := wm.x.instanceAndClass(c.window)
		if targetClass == clientClass {
			tag, _ := wm.findTag(c.tagMask)
			if wm.currTag != tag {
				wm.view(tag)
			}
			if wm.focusedClient != c {
				wm.focus(c)
			}
			return true
		}
	}

	return false
}

func (wm *windowManager) gotoWindowOrCreate(targetClass string, command string, args ...string) {
	found := wm.gotoWindow(targetClass)
	if found {
		return
	}

	go func() {
		cmd := exec.Command(command, args...)
		cmd.Start()

		for range 20 {
			time.Sleep(time.Millisecond * 100)
			found = wm.gotoWindow(targetClass)
			if found {
				return
			}
		}
	}()
}
