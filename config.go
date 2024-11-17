package main

const borderWidth = 2

type rule struct {
	class string
	tagId uint16
}

var rules = []rule{
	{"org.wezfu", 1 << 0},
	{"firefox", 1 << 1},
	{"TelegramDesktop", 1 << 2},
}
