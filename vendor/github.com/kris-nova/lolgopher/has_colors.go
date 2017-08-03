package lol

// #cgo LDFLAGS: -lncurses
// #include <ncurses.h>
//
import "C"

func hasColors() bool {
	C.initscr()
	C.start_color()
	hasColors := bool(C.has_colors())
	C.endwin()
	return hasColors
}
