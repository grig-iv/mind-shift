package main

import (
	"fmt"
	"sync"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

var (
	mu    sync.Mutex
	cache = make(map[string]xproto.Atom)
)

const (
	TransientName = "WM_TRANSIENT_FOR"
)

func getAtomProperty(conn *xgb.Conn, win xproto.Window, atomName string) (*xproto.GetPropertyReply, error) {
	atom, err := getAtom(conn, atomName)
	if err != nil {
		return nil, err
	}

	return xproto.GetProperty(conn, false, win, atom, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
}

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
