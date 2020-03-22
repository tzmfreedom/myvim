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
var frameY int

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
	frameY = 0
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
					buf := buffers[frameY+y]
					first := "" + buf[0:x]
					second := "" + buf[x:]
					newBuf := make([]string, y)
					copy(newBuf, buffers[0:y])
					last := buffers[y+1:]
					buffers = append(newBuf, first, second)
					buffers = append(buffers, last...)
					x = 0
					if y == height - 1 {
						frameY++
					} else {
						y++
					}
				}
			case termbox.KeySpace:
				buf := buffers[frameY+y]
				buffers[frameY+y] = buf[0:x] + " " + buf[x:]
				x++
			case termbox.KeyDelete, termbox.KeyBackspace, termbox.KeyBackspace2:
				if mode == ModeCommand {
					if len(commandBuffer) > 0 {
						commandBuffer = commandBuffer[:len(commandBuffer)-1]
					}
				} else {
					if x > 0 {
						buf := buffers[frameY+y]
						buffers[frameY+y] = buf[0:x-1] + buf[x:]
						x--
					}
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
					buf := buffers[frameY+y]
					buffers[frameY+y] = buf[0:x] + chr + buf[x:]
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
		if _, ok := err.(*os.PathError); ok {
			return []string{""}, nil
		}
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
		if deleteCommand {
			deleteWord()
			deleteCommand = false
		} else {
			word()
		}
	case "b":
		back()
	case "0", "^":
		x = 0
	case "$":
		x = len(buffers[frameY+y])
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
		if len(buffers[frameY+y]) < x {
			x = len(buffers[frameY+y])
		}
	}
}

func deleteWord() {
	buf := buffers[frameY+y]
	for i, b := range buf[x:] {
		if b == ' ' && x + i < len(buffers[frameY+y]) {
			if buffers[frameY+y][x+i+1] != ' ' {
				buffers[frameY+y] = buf[0:x] + buf[x + i + 1:]
				return
			}
		}
	}
	buffers[frameY+y] = buf[0:x]
}

func up() {
	if y + frameY == len(buffers) - height + 1 && frameY > 0 {
		frameY--
		return
	}
	if y > 0 {
		y--
		if len(buffers[frameY+y]) <= x {
			x = len(buffers[frameY+y])
		}
	}
}

func down() {
	if y == height - 2 {
		if y + frameY < len(buffers)-1 {
			frameY++
		}
	} else if len(buffers)-1 > frameY + y {
		y++
		if len(buffers[frameY+y]) < x {
			x = len(buffers[frameY+y])
		}
	}
}

func left() {
	if x > 0 {
		x--
	}
}

func right() {
	if len(buffers[frameY+y]) > x {
		x++
	}
}

func word() {
	for i, b := range buffers[frameY+y][x:] {
		if b == ' ' && x + i < len(buffers[frameY+y]) {
			if buffers[frameY+y][x+i+1] != ' ' {
				x += i + 1
				return
			}
		}
	}
	x = len(buffers[frameY+y])
}

func back() {
	if x == 0 {
		return
	}
	if x == len(buffers[frameY+y]) {
		x--
		return
	}
	for i := 0; i < x; i++ {
		if buffers[frameY+y][x-i] == ' ' {
			if buffers[frameY+y][x-i-1] != ' ' {
				x -= i + 1
				return
			}
		}
	}
	x = 0
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
	var frameBuffers []string
	if frameY+height < len(buffers) {
		frameBuffers = buffers[frameY:frameY+height]
	} else {
		frameBuffers = buffers[frameY:]
	}
	for by, buffer := range frameBuffers {
		if showNumber {
			fmt.Printf("\033[33m%02d: \033[39m", by+frameY)
		}
		for bx, b := range buffer {
			chr := string(b)
			if x == bx && y == by {
				drawCursorText(chr)
			} else {
				fmt.Print(chr)
			}
		}
		if y == by && len(buffers[frameY+y]) == x {
			drawCursorText(" ")
		}
		fmt.Print("\n")
	}
	if colonCommand {
		fmt.Printf("\033[%d;1H:%s", height-1, commandBuffer)
	}
	fmt.Printf("\033[%d;1H[%d:%d]", height, y+1, x+1)
	fmt.Printf("[%d:%d:%d:%d]", len(buffers), height, y, frameY)
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