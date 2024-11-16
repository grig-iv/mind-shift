// taken from https://github.com/BurntSushi/xgbutil/blob/0d543e91747d289a63a8698e021fdef1512fe766/xcursor/xcursor.go
package main

import "github.com/jezek/xgb/xproto"

func (x *xserver) CreateCursor(cursor uint16) (xproto.Cursor, error) {
	return x.CreateCursorExtra(cursor, 0, 0, 0, 0xffff, 0xffff, 0xffff)
}

func (x *xserver) CreateCursorExtra(cursor, foreRed, foreGreen,
	foreBlue, backRed, backGreen, backBlue uint16) (xproto.Cursor, error) {

	fontId, err := xproto.NewFontId(x.conn)
	if err != nil {
		return 0, err
	}

	cursorId, err := xproto.NewCursorId(x.conn)
	if err != nil {
		return 0, err
	}

	err = xproto.OpenFontChecked(x.conn, fontId,
		uint16(len("cursor")), "cursor").Check()
	if err != nil {
		return 0, err
	}

	err = xproto.CreateGlyphCursorChecked(x.conn, cursorId, fontId, fontId,
		cursor, cursor+1,
		foreRed, foreGreen, foreBlue,
		backRed, backGreen, backBlue).Check()
	if err != nil {
		return 0, err
	}

	err = xproto.CloseFontChecked(x.conn, fontId).Check()
	if err != nil {
		return 0, err
	}

	return cursorId, nil
}
