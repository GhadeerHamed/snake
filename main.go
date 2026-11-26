package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/gdamore/tcell"
)

/*
//
//  ███████╗███╗   ██╗ █████╗ ██╗  ██╗███████╗
//  ██╔════╝████╗  ██║██╔══██╗██║ ██╔╝██╔════╝
//  █████╗  ██╔██╗ ██║███████║█████╔╝ █████╗
//  ██╔══╝  ██║╚██╗██║██╔══██║██╔═██╗ ██╔══╝
//  ███████╗██║ ╚████║██║  ██║██║  ██╗███████╗
//  ╚══════╝╚═╝  ╚═══╝╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝
*/

const SnakeSymbol = 0x2588
const AppleSymbol = 0x25CF

const GameFrameWidth = 50
const GameFrameHeight = 20
const GameFrameSymbol = '|'

type Point struct {
	row, col int
}

type Snake struct {
	parts          []*Point
	velRow, velCol int
	symbol         rune
}

type Apple struct {
	point  *Point
	symbol rune
}

var screen tcell.Screen
var snake *Snake
var apple *Apple
var score int
var isGameOver bool
var IsGamePaused bool
var debugLog string

func main() {
	initScreen()
	initGameState()
	inputChan := initUserInput()

	for !isGameOver {
		handleUserInput(readInput(inputChan))
		UpdateState()
		DrawState()

		time.Sleep(100 * time.Millisecond)
	}

	screenWidth, screenHeight := screen.Size()
	PrintStringCenter(screenHeight/2, screenWidth/2, "Game Over!")
	PrintStringCenter(screenHeight/2+1, screenWidth/2, fmt.Sprintf("Your score: %d", score))
	screen.Show()
	time.Sleep(3 * time.Second)
	screen.Fini()
}

func DrawState() {
	if IsGamePaused {
		return
	}

	screen.Clear()
	PrintString(0, 0, debugLog)
	PrintGameFrame()
	PrintSnake()
	PrintApple()

	screen.Show()
}

func UpdateState() {
	if IsGamePaused {
		return
	}

	UpdateSnake()
	UpdateApple()
}

func UpdateSnake() {
	// add a new element
	head := GetSnakeHead()
	snake.parts = append(snake.parts, &Point{
		row: head.row + snake.velRow,
		col: head.col + snake.velCol,
	})

	if !AppleInsideSnake() {
		snake.parts = snake.parts[1:] // remove the tail
	} else {
		score++
	}

	if SnakeCollidedWithWall() || SnakeCollidedWithItself() {
		isGameOver = true
	}
}

func GetSnakeHead() *Point {
	return snake.parts[len(snake.parts)-1]
}

func SnakeCollidedWithItself() bool {
	head := GetSnakeHead()
	for _, p := range snake.parts[:len(snake.parts)-1] {
		if p.row == head.row && p.col == head.col {
			return true
		}
	}
	return false
}

func SnakeCollidedWithWall() bool {
	head := GetSnakeHead()
	if head.row < 1 || head.row > GameFrameHeight ||
		head.col < 1 || head.col > GameFrameWidth {
		return true
	}
	return false
}

func UpdateApple() {
	for AppleInsideSnake() {
		apple.point.row, apple.point.col =
			rand.Intn(GameFrameHeight-2)+1, rand.Intn(GameFrameWidth-2)+1
	}
}

func AppleInsideSnake() bool {
	for _, p := range snake.parts {
		if p.row == apple.point.row && p.col == apple.point.col {
			return true
		}
	}
	return false
}

func initUserInput() chan string {
	inputChan := make(chan string)
	go func() {
		for {
			switch ev := screen.PollEvent().(type) {
			case *tcell.EventKey:
				inputChan <- ev.Name()
			}
		}
	}()

	return inputChan
}

func initScreen() {
	var err error

	screen, err = tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if err = screen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	defStyle := tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorGreen)
	screen.SetStyle(defStyle)
}

func initGameState() {
	snake = &Snake{
		parts: []*Point{
			{row: 9, col: 3}, // tail
			{row: 8, col: 3},
			{row: 7, col: 3},
			{row: 6, col: 3},
			{row: 5, col: 3}, // head
		},
		velRow: -1,
		velCol: 0,
		symbol: SnakeSymbol,
	}
	apple = &Apple{
		point:  &Point{row: 15, col: 15},
		symbol: AppleSymbol,
	}
}

func readInput(inputChan chan string) string {
	var key string
	select {
	case key = <-inputChan:
	default:
		key = ""
	}

	return key
}

func handleUserInput(key string) {
	if key == "Rune[q]" {
		screen.Fini()
		os.Exit(0)
	} else if key == "Rune[p]" {
		IsGamePaused = !IsGamePaused
	} else if key == "Up" && snake.velRow != 1 {
		snake.velRow = -1
		snake.velCol = 0
	} else if key == "Down" && snake.velRow != -1 {
		snake.velRow = 1
		snake.velCol = 0
	} else if key == "Left" && snake.velCol != 1 {
		snake.velRow = 0
		snake.velCol = -1
	} else if key == "Right" && snake.velCol != -1 {
		snake.velRow = 0
		snake.velCol = 1
	}
}

func PrintGameFrame() {
	topLeftRow, topLeftCol := GetGameFrameTopLeft()
	width, height := GameFrameWidth+1, GameFrameHeight+1

	PrintUnFilledRect(topLeftRow, topLeftCol, width, height, GameFrameSymbol)
	// PrintUnFilledRect(topLeftRow+1, topLeftCol+1, GameFrameWidth, GameFrameHeight, '*')
}

func PrintSnake() {
	for _, p := range snake.parts {
		PrintFilledRectInGameFrame(p.row, p.col, 1, 1, snake.symbol)
	}
}

func PrintFilledRectInGameFrame(row, col, width, height int, ch rune) {
	r, c := GetGameFrameTopLeft()
	PrintFilledRect(row+r, col+c, width, height, ch)
}

func GetGameFrameTopLeft() (int, int) {
	screenWidth, screenHeight := screen.Size()
	r, c := screenHeight/2-GameFrameHeight/2, screenWidth/2-GameFrameWidth/2
	return r, c
}

func PrintApple() {
	PrintFilledRectInGameFrame(apple.point.row, apple.point.col, 1, 1, apple.symbol)
}

func PrintString(row, col int, str string) {
	for _, c := range str {
		PrintFilledRect(row, col, 1, 1, c)
		col += 1
	}
}

func PrintStringCenter(row, col int, str string) {
	col = col - len(str)/2
	PrintString(row, col, str)
}

func PrintFilledRect(row, col, width, height int, ch rune) {
	for r := 0; r < height; r++ {
		for c := 0; c < width; c++ {
			screen.SetContent(col+c, row+r, ch, nil, tcell.StyleDefault)
		}
	}
}

func PrintUnFilledRect(row, col, width, height int, ch rune) {
	// print first row
	for c := 0; c < width; c++ {
		screen.SetContent(col+c, row, ch, nil, tcell.StyleDefault)
	}

	// print first col, last col only for each row
	for r := 1; r < height-1; r++ {
		screen.SetContent(col, row+r, ch, nil, tcell.StyleDefault)
		screen.SetContent(col+width-1, row+r, ch, nil, tcell.StyleDefault)
	}

	// print last row
	for c := 0; c < width; c++ {
		screen.SetContent(col+c, row+height-1, ch, nil, tcell.StyleDefault)
	}
}
