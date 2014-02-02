package main

import (
	"github.com/nsf/termbox-go"
	"math/rand"
	"strconv"
	"time"
)

var update bool
var gameOver bool

type State int8

const (
	Safe  State = 0
	Bomb  State = 10
	Bomb2 State = 11
	Mask  State = 100
	Flag  State = 101
	Open  State = 102
)

type Board struct {
	Width   int
	Height  int
	Box     [][]State
	MaskBox [][]State
	CursorX int
	CursorY int
	Started bool
}

func (b *Board) setBomb(sx, sy int) {
	now := time.Now().UnixNano()
	rnd := rand.New(rand.NewSource(now))
	cnt := 0
	for {
		x := rnd.Intn(b.Width - 1)
		y := rnd.Intn(b.Height - 1)
		if x < sx-1 || x > sx+1 || y < sy-1 || y > sy+1 {
			if b.Box[y][x] == Safe {
				b.Box[y][x] = Bomb
				cnt++
			}
		}
		if cnt >= 10 {
			break
		}
	}
	for i := range b.Box {
		for j := range b.Box[i] {
			if b.Box[i][j] == Bomb {
				for m := i - 1; m <= i+1; m++ {
					for n := j - 1; n <= j+1; n++ {
						if m >= 0 && n >= 0 && m < b.Height && n < b.Width {
							if b.Box[m][n] != Bomb {
								b.Box[m][n]++
							}
						}
					}
				}
			}
		}
	}
}

func (b *Board) initBox() {
	ary := make([][]State, b.Height)
	for i := range ary {
		ary[i] = make([]State, b.Width)
	}
	b.Box = ary

	maskAry := make([][]State, b.Height)
	for i := range maskAry {
		maskAry[i] = make([]State, b.Width)
		for j := range maskAry[i] {
			maskAry[i][j] = Mask
		}
	}
	b.MaskBox = maskAry
}

func (b *Board) toggleFlag(x, y int) {
	if b.MaskBox[y][x] != Open {
		b.MaskBox[y][x] ^= 1
		update = true
	}
}

func (b *Board) expand(x, y int) {
	for i := y - 1; i <= y+1; i++ {
		if i < 0 || i >= b.Height {
			continue
		}
		for j := x - 1; j <= x+1; j++ {
			if j < 0 || j >= b.Width {
				continue
			}
			if x != j || y != i {
				if b.MaskBox[i][j] == Mask {
					b.MaskBox[i][j] = Open
					if b.Box[i][j] == Safe {
						b.expand(j, i)
					}
				}
			}
		}
	}
}

func (b *Board) open(x, y int) {
	if !b.Started {
		b.setBomb(x, y)
		b.Started = true
	}
	if b.MaskBox[y][x] == Mask {
		b.MaskBox[y][x] = Open
		if b.Box[y][x] == Safe {
			b.expand(x, y)
		} else if b.Box[y][x] == Bomb {
			b.Box[y][x] = Bomb2
			b.openBomb()
			gameOver = true
		}
		update = true
	}
}

func (b *Board) openBomb() {
	for i := range b.Box {
		for j := range b.Box[i] {
			if b.Box[i][j] == Bomb && b.MaskBox[i][j] == Mask {
				b.MaskBox[i][j] = Open
			}
		}
	}
}

func NewBoard(w, h int) *Board {
	b := new(Board)
	b.Width = w
	b.Height = h
	b.Started = false
	b.initBox()
	return b
}

func setString(x, y int, s string, fc, bc termbox.Attribute) {
	for i, chr := range s {
		termbox.SetCell(x+i, y, chr, fc, bc)
	}
}

func setStringV(x, y int, s string, fc, bc termbox.Attribute) {
	for i, chr := range s {
		termbox.SetCell(x, y+i, chr, fc, bc)
	}
}

func draw(b *Board) {
	if update {
		for i := range b.Box {
			for j, v := range b.Box[i] {
				if b.MaskBox[i][j] != Open {
					if b.MaskBox[i][j] == Mask {
						termbox.SetCell(j, i, ' ',
							termbox.ColorWhite, termbox.ColorWhite)
					} else {
						termbox.SetCell(j, i, 'F',
							termbox.ColorBlack, termbox.ColorWhite)
					}
				} else {
					if v == Bomb {
						termbox.SetCell(j, i, '@',
							termbox.ColorWhite, termbox.ColorDefault)
					} else if v == Bomb2 {
						termbox.SetCell(j, i, '@',
							termbox.ColorWhite, termbox.ColorRed)
					} else if v == Safe {
						termbox.SetCell(j, i, ' ',
							termbox.ColorWhite, termbox.ColorDefault)
					} else {
						s := strconv.Itoa(int(b.Box[i][j]))
						termbox.SetCell(j, i, rune(s[0]),
							termbox.ColorWhite, termbox.ColorDefault)
					}
				}
			}
		}
	}
	if !gameOver {
		termbox.SetCursor(b.CursorX, b.CursorY)
	} else {
		termbox.HideCursor()
	}
	termbox.Flush()
	update = false
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)

	board := NewBoard(9, 9)
	update = true
	gameOver = false
	setString(12, 0, "Move: j/k/h/l", termbox.ColorWhite, termbox.ColorDefault)
	setString(12, 1, "Open: space", termbox.ColorWhite, termbox.ColorDefault)
	setString(12, 2, "Flag: f", termbox.ColorWhite, termbox.ColorDefault)
	setString(12, 3, "Quit: Esc", termbox.ColorWhite, termbox.ColorDefault)

	tick := time.Tick(100 * time.Millisecond)
	stop := make(chan int)

	go func(b *Board) {
	stop:
		for {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				if ev.Key == termbox.KeyEsc {
					stop <- 1
					break stop
				}
				if ev.Ch == 'j' {
					if b.CursorY < b.Height-1 {
						b.CursorY++
					}
				}
				if ev.Ch == 'k' {
					if b.CursorY > 0 {
						b.CursorY--
					}
				}
				if ev.Ch == 'l' {
					if b.CursorX < b.Width-1 {
						b.CursorX++
					}
				}
				if ev.Ch == 'h' {
					if b.CursorX > 0 {
						b.CursorX--
					}
				}
				if ev.Ch == 'f' && !gameOver {
					b.toggleFlag(b.CursorX, b.CursorY)
				}
				if ev.Key == termbox.KeySpace && !gameOver {
					b.open(b.CursorX, b.CursorY)
				}
			case termbox.EventError:
				panic(ev.Err)
			}
		}
	}(board)

loop:
	for {
		select {
		case <-tick:
			draw(board)
		case <-stop:
			break loop
		}
	}
}
