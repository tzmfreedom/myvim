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
var width, height int
var mode int
var deleteCommand bool
var colonCommand bool
var commandBuffer string
var showNumber bool
var filename string

const (
	ModeCommand = iota
	ModeEdit
)

type Command struct {
	Write bool
	Quite bool
}

var buffers []string

func main() {
	x = 0
	y = 0
	deleteCommand = false
	showNumber = true
	filename = os.Args[1]
	colonCommand = false
	commandBuffer = ""

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
	width, height = termbox.Size()

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
				if colonCommand {
					command := parseColonCommand(commandBuffer)
					if command.Write {
						save(filename)
					}
					if command.Quite {
						return
					}

					colonCommand = false
					commandBuffer = ""
				} else {
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
				}
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
					if colonCommand {
						if string(ev.Ch) != ":" {
							commandBuffer += string(ev.Ch)
						}
					} else {
						if string(ev.Ch) == ":" {
							colonCommand = true
							commandBuffer = ""
						} else {
							handleCommand(ev)
						}
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
	case "w":
		save(filename)
	case "0", "^":
		x = 0
	case "$":
		x = len(buffers[y])
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

func parseColonCommand(buffer string) *Command {
	command := &Command{}
	if strings.ContainsRune(buffer, 'w') {
		command.Write = true
	}
	if strings.ContainsRune(buffer, 'q') {
		command.Quite = true
	}
	return command
}

func save(filename string) {
	ioutil.WriteFile(filename, []byte(strings.Join(buffers, "\n")), 0644)
}

func draw() {
	fmt.Print("\033[2J")
	fmt.Print("\r")
	fmt.Print("\033[;H")
	for by, buffer := range buffers {
		if showNumber {
			fmt.Printf("\033[33m%02d: \033[39m", by)
		}
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
	if colonCommand {
		fmt.Printf("\033[%d;1H:%s", height-1, commandBuffer)
	}
	fmt.Printf("\033[%d;1H[%d:%d]", height, y+1, x+1)
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