package main

import (
	"log"

	"github.com/jezek/xgb/xproto"
)

func main() {
	wm := newWindowManager()

	err := wm.x.checkOtherWm()
	if err != nil {
		log.Fatal("Another window manager might aleady be running:", err)
	}

	log.Println("Starting mind-shift")

	kbm, err := newKeyboardManager(wm)
	if err != nil {
		log.Fatal("Failed to creat keyboard manager:", err)
	}

	kbm.grabKeys()

	wm.scan()
	wm.view(wm.currTag)

	wm.isRunning = true
	for wm.isRunning {
		ev, xerr := wm.x.conn.WaitForEvent()
		if ev == nil && xerr == nil {
			log.Println("Both event and error are nil. Exiting...")
			return
		}

		if xerr != nil {
			log.Printf("Error: %s\n", xerr)
		}

		if ev == nil {
			continue
		}

		switch v := ev.(type) {
		case xproto.KeyPressEvent:
			log.Println("-> KeyPressEvent")
			kbm.onKeyPress(v)
		case xproto.MapRequestEvent:
			log.Println("-> MapRequestEvent")
			wm.onMapRequest(v)
		case xproto.ConfigureNotifyEvent:
			wm.onConfigureNotify(v)
		case xproto.ConfigureRequestEvent:
			log.Println("-> ConfigureRequestEvent")
			wm.onConfigureRequest(v)
		case xproto.DestroyNotifyEvent:
			log.Println("-> DestroyNotifyEvent")
			wm.onDestroyNotify(v)
		case xproto.ButtonPressEvent:
			log.Println("-> ButtonPressEvent")
			wm.onButtonPressEvent(v)
		case xproto.ClientMessageEvent:
			wm.onClientMessageEvent(v)
		case xproto.CreateNotifyEvent:
		case xproto.MapNotifyEvent:
		case xproto.MotionNotifyEvent:
			continue
		default:
			// log.Printf("-> [skip] %T\n", v)
		}
	}

	wm.cleanup()
}
