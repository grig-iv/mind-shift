package main

import (
	"log"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type keyboardManager struct {
	conn             *xgb.Conn
	setup            *xproto.SetupInfo
	gestureToCommand map[gesture]command
}

type gesture struct {
	mods uint16
	code xproto.Keycode
}

type keyBinding struct {
	mods    uint16
	keysym  xproto.Keysym
	command command
}

type command func()

func getKeybindings(wm *windowManager) []keyBinding {
	return []keyBinding{
		{xproto.ModMask4 | xproto.ModMask1 | xproto.ModMaskControl, keysyms["q"], wm.quit},
		{xproto.ModMask4 | xproto.ModMask1, keysyms["q"], wm.killClient},
		{xproto.ModMask4 | xproto.ModMaskControl, keysyms["Prior"], wm.gotoPrevTag},
		{xproto.ModMask4 | xproto.ModMaskControl, keysyms["Next"], wm.gotoNextTag},
		{xproto.ModMask4 | xproto.ModMaskShift | xproto.ModMaskControl, keysyms["Prior"], wm.moveToPrevTag},
		{xproto.ModMask4 | xproto.ModMaskShift | xproto.ModMaskControl, keysyms["Next"], wm.moveToNextTag},
	}
}

func newKeyboardManager(wm *windowManager) (*keyboardManager, error) {
	gestureToCommand, err := getGestureToCommand(wm)
	if err != nil {
		return nil, err
	}

	kbm := &keyboardManager{
		conn:             wm.x.conn,
		setup:            wm.setup,
		gestureToCommand: gestureToCommand,
	}

	kbm.grabKeys()

	return kbm, nil
}

func getGestureToCommand(wm *windowManager) (map[gesture]command, error) {
	gestureToCommand := make(map[gesture]command)

	minCode := wm.setup.MinKeycode
	maxCode := wm.setup.MaxKeycode

	mapping, err := xproto.GetKeyboardMapping(wm.x.conn, minCode, byte(maxCode-minCode+1)).Reply()
	if err != nil {
		return nil, err
	}
	if len(mapping.Keysyms) == 0 {
		log.Println("no key syms")
	}

	keybindings := getKeybindings(wm)
	for code := minCode; code >= minCode && code <= maxCode; code++ {
		for _, kb := range keybindings {
			index := int(code-minCode) * int(mapping.KeysymsPerKeycode)
			if kb.keysym == mapping.Keysyms[index] {
				gesture := gesture{kb.mods, code}
				gestureToCommand[gesture] = kb.command
			}
		}
	}

	return gestureToCommand, nil
}

func (kbm *keyboardManager) grabKeys() {
	xproto.UngrabKey(kbm.conn, xproto.GrabAny, kbm.xRoot(), xproto.ModMaskAny)

	ignoredModes := []uint16{0, xproto.ModMaskLock} // add numlock

	for gesture := range kbm.gestureToCommand {
		for _, ignore := range ignoredModes {
			xproto.GrabKey(
				kbm.conn,
				true,
				kbm.xRoot(),
				gesture.mods|ignore,
				gesture.code,
				xproto.GrabModeAsync,
				xproto.GrabModeAsync,
			)
		}
	}
}

func (kbm *keyboardManager) onKeyPress(event xproto.KeyPressEvent) {
	if command, ok := kbm.gestureToCommand[gesture{event.State, event.Detail}]; ok {
		command()
	}
}

func (kbm *keyboardManager) xRoot() xproto.Window {
	return kbm.setup.DefaultScreen(kbm.conn).Root
}
