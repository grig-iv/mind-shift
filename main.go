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

type windowManager struct {
	clients []client
}

type client struct {
	win  xproto.Window
	geom geometry
	tags []tag
}

type geometry struct {
	x, y          int
	width, height uint
}

type tag uint8

var (
	wm    windowManager
	conn  *xgb.Conn
	setup *xproto.SetupInfo
	root  xproto.Window
)

func main() {
	c, err := xgb.NewConn()
	if err != nil {
		log.Fatal("Failed to connect to X server:", err)
	}
	defer c.Close()

	conn = c
	setup = xproto.Setup(conn)
	root = setup.DefaultScreen(conn).Root
	wm = windowManager{make([]client, 0)}

	checkOtherWm()
	scan()
	arrange()

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

		if i > 30 {
			return
		}
		i += 1
	}
}

func checkOtherWm() {
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

func scan() {
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

		if winAttrs.MapState == xproto.MapStateUnmapped {
			continue
		}

		transient, err := getAtomProperty(conn, win, TransientName)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if len(transient.Value) != 0 {
			fmt.Println("Fount transient window: ", transient)
			continue
		}

		manage(win)
	}

}

func manage(win xproto.Window) {
	drw := xproto.Drawable(win)
	geom, err := xproto.GetGeometry(conn, drw).Reply()
	if err != nil {
		fmt.Println("GetGeometry error: ", err)
	}
	fmt.Println(geom)

	client := client{
		win:  win,
		tags: []tag{1},
		geom: geometry{
			x:      int(geom.X),
			y:      int(geom.Y),
			width:  uint(geom.Width),
			height: uint(geom.Height),
		},
	}

	wm.clients = append(wm.clients, client)
}

func arrange() {
	const (
		clientPadding = 8
		screenPadding = 16
		masterRatio   = 0.5
	)

	if len(wm.clients) == 0 {
		return
	}

	screen := setup.DefaultScreen(conn)
	screenGeom := geometry{
		x:      screenPadding,
		y:      screenPadding,
		width:  uint(screen.WidthInPixels) - (screenPadding * 2),
		height: uint(screen.HeightInPixels) - (screenPadding * 2),
	}

	if len(wm.clients) == 1 {
		changeGeometry(wm.clients[0], screenGeom)
	} else {
		masterGeom := geometry{
			x:      screenGeom.x,
			y:      screenGeom.y,
			width:  uint(float64(screenGeom.width) * masterRatio),
			height: screenGeom.height,
		}
		changeGeometry(wm.clients[0], masterGeom)

		stackGeom := geometry{
			x:      masterGeom.x + int(masterGeom.width) + clientPadding,
			y:      screenGeom.y,
			width:  screenGeom.width - masterGeom.width - clientPadding,
			height: screenGeom.height,
		}

		totalStackClients := len(wm.clients) - 1
		clientHeight := (int(screenGeom.height) / totalStackClients) - (clientPadding * (totalStackClients - 1))
		for i, c := range wm.clients[1:] {
			y := screenGeom.y + clientHeight*i + clientPadding*i
			clientGeom := geometry{
				x:      stackGeom.x,
				y:      y,
				width:  stackGeom.width,
				height: uint(clientHeight),
			}
			changeGeometry(c, clientGeom)
		}
	}
	// get monitor dementions
	// calculate layout
	// applay paddings
}

func changeGeometry(client client, newGeom geometry) {
	newGeom = geometry{
		x:      newGeom.x,
		y:      newGeom.y,
		width:  max(newGeom.width, 1),
		height: max(newGeom.height, 1),
	}

	vals := []uint32{
		uint32(newGeom.x),
		uint32(newGeom.y),
		uint32(newGeom.width),
		uint32(newGeom.height),
	}

	xproto.ConfigureWindow(conn, client.win,
		xproto.ConfigWindowX|xproto.ConfigWindowY|
			xproto.ConfigWindowWidth|xproto.ConfigWindowHeight, vals)

}
