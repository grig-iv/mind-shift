package main

import (
	"log"

	"github.com/jezek/xgb/xproto"
)

func (wm *windowManager) setup() {
	wm.x.checkOtherWm()
	wm.scan()
	wm.view(wm.currTag)

	cursor, err := wm.x.CreateCursor(leftPtrCursor)
	if err != nil {
		log.Println(err)
	} else {
		xproto.ChangeWindowAttributes(
			wm.x.conn,
			wm.x.root,
			xproto.CwCursor,
			[]uint32{uint32(cursor)},
		)
	}
}

func (wm *windowManager) scan() {
	queryTree, err := xproto.QueryTree(wm.x.conn, wm.x.root).Reply()
	if err != nil {
		log.Println(err)
	}

	for _, win := range queryTree.Children {
		winAttrs, err := xproto.GetWindowAttributes(wm.x.conn, win).Reply()
		if err != nil {
			log.Println(err)
			continue
		}

		if winAttrs.OverrideRedirect {
			continue
		}

		if winAttrs.MapState == xproto.MapStateUnmapped {
			continue
		}

		transient, err := wm.x.atomProperty(win, WMTransientName)
		if err != nil {
			log.Println(err)
			continue
		}
		if len(transient.Value) != 0 {
			log.Println("Fount transient window: ", transient)
			continue
		}

		_, class := wm.x.instanceAndClass(win)

		if wm.bar.isBar(class) {
			wm.bar.register(win)
			wm.x.changeGeometry(win, wm.bar.geom)
			continue
		}

		wm.manageClient(win, class)
	}
}
