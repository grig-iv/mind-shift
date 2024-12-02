package main

import (
	"log"

	"github.com/grig-iv/mind-shift/x"
	"github.com/jezek/xgb/xproto"
)

func (wm *wm) setup() {
	x.CheckOtherWm()
	wm.scan()
	wm.view(wm.currTag)

	cursor, err := x.CreateCursor(x.LeftPtrCursor)
	if err != nil {
		log.Println(err)
	} else {
		xproto.ChangeWindowAttributes(
			x.Conn,
			x.Root,
			xproto.CwCursor,
			[]uint32{uint32(cursor)},
		)
	}
}

func (wm *wm) scan() {
	queryTree, err := xproto.QueryTree(x.Conn, x.Root).Reply()
	if err != nil {
		log.Println(err)
	}

	for _, win := range queryTree.Children {
		winAttrs, err := xproto.GetWindowAttributes(x.Conn, win).Reply()
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

		transient, err := x.AtomProperty(win, x.WMTransientName)
		if err != nil {
			log.Println(err)
			continue
		}
		if len(transient.Value) != 0 {
			log.Println("Fount transient window: ", transient)
			continue
		}

		_, class := x.InstanceAndClass(win)

		if wm.bar.isBar(class) {
			wm.bar.register(win)
			x.ChangeGeometry(win, wm.bar.geom)
			continue
		}

		wm.manage(win, class)
	}
}
