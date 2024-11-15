package main

import (
	"fmt"
	"log"
)

const borderWidth = 2

func main() {
	wm := newWindowManager()

	wm.x.checkOtherWm()

	log.Println("Starting mind-shift")

	kbm := newKeyboardManager(wm)

	wm.scan()
	wm.view(wm.currTag)

	fmt.Print("\n\n")

	wm.loop(kbm)
	wm.cleanup()
}
