package main

import (
	"log"

	"github.com/grig-iv/mind-shift/socket"
	"github.com/grig-iv/mind-shift/x"
	"github.com/jezek/xgb/xproto"
)

type keyboardManager struct {
	setup *xproto.SetupInfo
	wm    *windowManager

	gestureToCommand map[gesture]socket.Cmd
}

type gesture struct {
	mods uint16
	code xproto.Keycode
}

type keyBinding struct {
	mods   uint16
	keysym xproto.Keysym
	cmd    socket.Cmd
}

type command func()

func getKeybindings() []keyBinding {
	return []keyBinding{
		{xproto.ModMask4, keysyms["t"], socket.GoToWinOrSpawnCmd{Class: "org.wezfu", SpanCmd: "wezterm"}},
		{xproto.ModMask4, keysyms["f"], socket.GoToWinOrSpawnCmd{Class: "firefox", SpanCmd: "firefox"}},
		{xproto.ModMask4, keysyms["s"], socket.GoToWinOrSpawnCmd{Class: "TelegramDesktop", SpanCmd: "telegram-desktop"}},
	}
}

func newKeyboardManager(wm *windowManager) *keyboardManager {
	gestureToCommand, err := getGestureToCommand()
	if err != nil {
		log.Fatal("Failed to creat keyboard manager:", err)
	}

	kbm := &keyboardManager{
		wm:               wm,
		gestureToCommand: gestureToCommand,
	}

	kbm.grabKeys()

	return kbm
}

func getGestureToCommand() (map[gesture]socket.Cmd, error) {
	gestureToCommand := make(map[gesture]socket.Cmd)

	minCode := x.Setup.MinKeycode
	maxCode := x.Setup.MaxKeycode

	mapping, err := xproto.GetKeyboardMapping(x.Conn, minCode, byte(maxCode-minCode+1)).Reply()
	if err != nil {
		return nil, err
	}
	if len(mapping.Keysyms) == 0 {
		log.Println("no key syms")
	}

	keybindings := getKeybindings()
	for code := minCode; code >= minCode && code <= maxCode; code++ {
		for _, kb := range keybindings {
			index := int(code-minCode) * int(mapping.KeysymsPerKeycode)
			if kb.keysym == mapping.Keysyms[index] {
				gesture := gesture{kb.mods, code}
				gestureToCommand[gesture] = kb.cmd
			}
		}
	}

	return gestureToCommand, nil
}

func (kbm *keyboardManager) grabKeys() {
	xproto.UngrabKey(x.Conn, xproto.GrabAny, x.Root, xproto.ModMaskAny)

	ignoredModes := []uint16{0, xproto.ModMaskLock} // add numlock

	for gesture := range kbm.gestureToCommand {
		for _, ignore := range ignoredModes {
			xproto.GrabKey(
				x.Conn,
				true,
				x.Root,
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
		kbm.wm.eval(command)
	}
}
