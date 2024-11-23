package main

const borderWidth = 2

type rule struct {
	class string
	tagId uint16
}

const (
	weztermClass  = "org.wezfu"
	firefoxClass  = "firefox"
	telegramClass = "TelegramDesktop"
)

var rules = []rule{
	{weztermClass, 1 << 0},
	{firefoxClass, 1 << 1},
	{telegramClass, 1 << 2},
}

func onStartup(wm *windowManager) {
	wm.spawnIfNotExist(weztermClass, "wezterm", "-e", "tmuxp load main -y")
	wm.spawnIfNotExist(firefoxClass, "firefox")
	wm.spawnIfNotExist(telegramClass, "telegram-desktop")
}
