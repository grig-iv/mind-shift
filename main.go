package main

import (
	"fmt"
	"log"
)

const borderWidth = 2

func main() {
	wm := newWindowManager()

	log.Println("Starting mind-shift")

	wm.setup()

	fmt.Print("\n\n")

	wm.loop()
	wm.cleanup()
}
