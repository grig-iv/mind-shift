package main

import (
	"fmt"
	"log"
)

func main() {
	wm := newWindowManager()

	log.Println("Starting mind-shift")

	wm.setup()
	onStartup(wm)

	fmt.Print("\n\n")

	wm.loop()
	wm.cleanup()
}
