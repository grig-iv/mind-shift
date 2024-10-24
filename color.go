package main

import (
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

type color struct {
	r, g, b uint16
}

type colorTable map[color]uint32

var (
	normBorder  = color{68, 68, 68}
	focusBorder = color{0, 85, 119}
)

func createColorTable(conn *xgb.Conn, colormap xproto.Colormap) (map[color]uint32, error) {
	table := make(map[color]uint32)

	colors := []color{normBorder, focusBorder}

	colorToCookie := make(map[color]xproto.AllocColorCookie)
	for _, color := range colors {
		colorToCookie[color] = xproto.AllocColor(conn, colormap, color.r, color.g, color.b)
	}

	for color, cookie := range colorToCookie {
		replay, err := cookie.Reply()
		if err != nil {
			return nil, err
		}

		table[color] = replay.Pixel
	}

	return table, nil
}
