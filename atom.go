package main

import (
	"fmt"
	"sync"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

var (
	mu    sync.Mutex
	cache = make(map[string]xproto.Atom)
)

const (
	WMTransientName = "WM_TRANSIENT_FOR"
	WMProtocols     = "WM_PROTOCOLS"
	WMDelete        = "WM_DELETE_WINDOW"
	WMState         = "WM_STATE"
	WMTakeFocus     = "WM_TAKE_FOCUS"

	NetActiveWindow       = "_NET_ACTIVE_WINDOW"
	NetSupported          = "_NET_SUPPORTED"
	NetWMName             = "_NET_WM_NAME"
	NetWMState            = "_NET_WM_STATE"
	NetWMCheck            = "_NET_SUPPORTING_WM_CHECK"
	NetWMFullscreen       = "_NET_WM_STATE_FULLSCREEN"
	NetWMWindowType       = "_NET_WM_WINDOW_TYPE"
	NetWMWindowTypeDialog = "_NET_WM_WINDOW_TYPE_DIALOG"
	NetClientList         = "_NET_CLIENT_LIST"
)

func getAtomProperty(conn *xgb.Conn, win xproto.Window, atomName string) (*xproto.GetPropertyReply, error) {
	atom, err := getAtom(conn, atomName)
	if err != nil {
		return nil, err
	}

	return xproto.GetProperty(conn, false, win, atom, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
}

//
// func setAtomProperty(conn *xgb.Conn) {
// 	xproto.ChangeProperty(conn,		)
// }

func getAtom(conn *xgb.Conn, atomName string) (xproto.Atom, error) {
	atom, ok := findInCache(atomName)
	if ok {
		return atom, nil
	}

	reply, err := xproto.InternAtom(conn, false, uint16(len(atomName)), atomName).Reply()
	if err != nil {
		return 0, fmt.Errorf("Atom: Error interning atom '%s': %s", atomName, err)
	}

	addToCache(atomName, reply.Atom)

	return reply.Atom, nil
}

func findInCache(atomName string) (xproto.Atom, bool) {
	mu.Lock()
	defer mu.Unlock()

	atom, ok := cache[atomName]

	return atom, ok
}

func addToCache(atomName string, atom xproto.Atom) {
	mu.Lock()
	defer mu.Unlock()

	cache[atomName] = atom
}
