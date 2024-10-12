// Example create-window shows how to create a window, map it, resize it,
// and listen to structure and key events (i.e., when the window is resized
// by the window manager, or when key presses/releases are made when the
// window has focus). The events are printed to stdout.
package main

import (
	"fmt"
	"log"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

var (
	setup *xproto.SetupInfo
	root  xproto.Window
)

func checkOtherWm(conn *xgb.Conn) {
	err := xproto.ChangeWindowAttributesChecked(conn, root, xproto.CwEventMask,
		[]uint32{xproto.EventMaskSubstructureRedirect |
			xproto.EventMaskSubstructureNotify |
			xproto.EventMaskButtonPress |
			// xproto.EventMaskPointerMotion |
			xproto.EventMaskEnterWindow |
			xproto.EventMaskLeaveWindow |
			xproto.EventMaskStructureNotify |
			xproto.EventMaskPropertyChange}).Check()

	if err != nil {
		log.Fatal("Another window manager might already be running:", err)
	}
}

func scan(conn *xgb.Conn) {
	queryTree, err := xproto.QueryTree(conn, root).Reply()
	if err != nil {
		log.Fatal(err)
	}

	for _, win := range queryTree.Children {
		winAttrs, err := xproto.GetWindowAttributes(conn, win).Reply()
		if err != nil {
			fmt.Println(err)
			continue
		}

		if winAttrs.OverrideRedirect {
			continue
		}

		t, err := getAtomProperty(conn, win, TransientName)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if len(t.Value) != 0 {
			fmt.Println("Fount transient window: ", t)
			continue
		}

		fmt.Println(t)
	}

}

func main() {
	conn, err := xgb.NewConn()
	if err != nil {
		log.Fatal("Failed to connect to X server:", err)
	}
	defer conn.Close()

	setup = xproto.Setup(conn)
	root = setup.DefaultScreen(conn).Root

	checkOtherWm(conn)
	scan(conn)

	return
	i := 0
	for {
		ev, xerr := conn.WaitForEvent()
		if ev == nil && xerr == nil {
			fmt.Println("Both event and error are nil. Exiting...")
			return
		}

		if xerr != nil {
			fmt.Printf("Error: %s\n", xerr)
		}

		switch v := ev.(type) {
		case xproto.KeyReleaseEvent:
			fmt.Printf("Button pressed!\n")
			return
		default:
			fmt.Printf("I don't know about type %T!\n", v)
		}

		fmt.Println(fmt.Sprint(i))
		if i > 5 {
			return
		}
		i += 1
	}
}
