package main

import (
	"bufio"
	"fmt"
	"github.com/k0kubun/pp"
	"github.com/nsf/termbox-go"
	"io/ioutil"
	"os"
	"strings"
)
var x, y int
var mode int
var deleteCommand bool

const (
	ModeCommand = iota
	ModeEdit
)

var buffers []string

func main() {
	x = 0
	y = 0
	deleteCommand = false
	filename := os.Args[1]
	var err error
	buffers, err = readFile(filename)
	if err != nil {
		panic(err)
	}
	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	for {
		draw()
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				if mode == ModeEdit {
					mode = ModeCommand
				}
			case termbox.KeyCtrlS:
				save(filename)
			case termbox.KeyArrowUp:
				up()
			case termbox.KeyArrowDown:
				down()
			case termbox.KeyArrowLeft:
				left()
			case termbox.KeyArrowRight:
				right()
			case termbox.KeyEnter:
				buf := buffers[y]
				first := "" + buf[0:x]
				second := "" + buf[x:]
				newBuf := make([]string, y)
				copy(newBuf, buffers[0:y])
				last := buffers[y+1:]
				buffers = append(newBuf, first, second)
				buffers = append(buffers, last...)
				x = 0
				y++
			case termbox.KeySpace:
				buf := buffers[y]
				buffers[y] = buf[0:x] + " " + buf[x:]
				x++
			case termbox.KeyDelete, termbox.KeyBackspace, termbox.KeyBackspace2:
				if x > 0 {
					buf := buffers[y]
					buffers[y] = buf[0:x-1] + buf[x:]
					x--
				}
			default:
				if mode == ModeCommand {
					if string(ev.Ch) == "w" {
						save(filename)
					} else if string(ev.Ch) == "q" {
						return
					} else {
						handleCommand(ev)
					}
				} else {
					chr := string(ev.Ch)
					buf := buffers[y]
					buffers[y] = buf[0:x] + chr + buf[x:]
					x++
				}
			}
		default:
		}
	}
}

func readFile(filename string) ([]string, error) {
	fp, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)

	for scanner.Scan() {
		buffers = append(buffers, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return buffers, nil
}

func handleCommand(ev termbox.Event) {
	switch string(ev.Ch) {
	case "h":
		left()
	case "i":
		mode = ModeEdit
	case "j":
		down()
	case "k":
		up()
	case "l":
		right()
	case "d":
		if deleteCommand {
			deleteLine()
		}
		deleteCommand = !deleteCommand
	}
}

func deleteLine() {
	if len(buffers) == 1 {
		buffers = []string{""}
	} else {
		newBuf := make([]string, y)
		copy(newBuf, buffers[0:y])
		buffers = append(newBuf, buffers[y+1:]...)
		if len(buffers) <= y {
			y = len(buffers)-1
		}
		if len(buffers[y]) < x {
			x = len(buffers[y])
		}
	}
}

func up() {
	if y > 0 {
		y--
		if len(buffers[y]) <= x {
			x = len(buffers[y])
		}
	}
}

func down() {
	if len(buffers)-1 > y {
		y++
		if len(buffers[y]) < x {
			x = len(buffers[y])
		}
	}
}

func left() {
	if x > 0 {
		x--
	}
}

func right() {
	if len(buffers[y]) > x {
		x++
	}
}

func save(filename string) {
	ioutil.WriteFile(filename, []byte(strings.Join(buffers, "\n")), 0644)
}

func draw() {
	fmt.Print("\033[2J")
	fmt.Print("\r")
	fmt.Print("\033[;H")
	for by, buffer := range buffers {
		for bx, b := range buffer {
			chr := string(b)
			if x == bx && y == by {
				drawCursorText(chr)
			} else {
				fmt.Print(chr)
			}
		}
		if y == by && len(buffers[y]) == x {
			drawCursorText(" ")
		}
		fmt.Print("\n")
	}
}

func drawCursorText(chr string) {
	if mode == ModeCommand {
		fmt.Print("\033[43m" + chr + "\033[49m")
	} else {
		fmt.Print("\033[42m" + chr + "\033[49m")
	}
}

func debug(args ...interface{}) {
	pp.Println(args...)
}