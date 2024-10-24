package main

type tag struct {
	id        uint16
	currLaout layout
}

func newTagFromIndex(index uint, layout layout) tag {
	return tag{1 << index, layout}
}

func (tag tag) index() int {
	index := 0
	for tag.id>>index != 1 {
		index += 1
	}
	return index
}
