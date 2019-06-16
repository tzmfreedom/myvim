package main

import (
	"fmt"
	"github.com/k0kubun/pp"
	"github.com/nsf/termbox-go"
)
var timer int
var x, y int
var phase int

var buffers []string

func main() {
	x = 0
	y = 0
	buffers = []string{""}
	err := termbox.Init()
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
				return
			case termbox.KeyArrowUp:
				if y > 0 {
					y--
					if len(buffers[y]) <= x {
						x = len(buffers[y])
					}
				}
			case termbox.KeyArrowDown:
				if len(buffers)-1 > y {
					y++
					if len(buffers[y]) < x {
						x = len(buffers[y])
					}
				}
			case termbox.KeyArrowLeft:
				if x > 0 {
					x--
				}
			case termbox.KeyArrowRight:
				if len(buffers[y]) > x {
					x++
				}
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
				chr := string(ev.Ch)
				buf := buffers[y]
				buffers[y] = buf[0:x] + chr + buf[x:]
				x++
			}
		default:
		}
	}
}

func draw() {
	//debug(buffer)
	fmt.Print("\033[2J")
	fmt.Print("\r")
	fmt.Print("\033[;H")
	for by, buffer := range buffers {
		for bx, b := range buffer {
			chr := string(b)
			if x == bx && y == by {
				fmt.Print("\033[42m" + chr + "\033[49m")
			} else {
				fmt.Print(chr)
			}
		}
		if y == by && len(buffers[y]) == x {
			fmt.Print("\033[42m \033[49m")
		}
		fmt.Print("\n")
	}
}

func debug(args ...interface{}) {
	pp.Println(args...)
}