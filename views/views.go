package views

//
// The "views" package creates a layer that avoids to write logics tightly coupled with "termbox".
//

import (
	"fmt"
	"github.com/kjirou/gRPC-sample-net-game/utils"
	"github.com/nsf/termbox-go"
)

type ScreenCellProps struct {
	Symbol          rune
	Foreground termbox.Attribute
	Background termbox.Attribute
}

type screenCell struct {
	symbol          rune
	foreground termbox.Attribute
	background termbox.Attribute
}

func (screenCell *screenCell) render(props *ScreenCellProps) {
	screenCell.symbol = props.Symbol
	screenCell.foreground = props.Foreground
	screenCell.background = props.Background
}

type screenText struct {
	Position *utils.MatrixPosition
	// ASCII only. Line breaks are not allowed.
	Text string
	Foreground termbox.Attribute
}

type ScreenProps struct {
	FieldCells [][]*ScreenCellProps
	FloorNumber int
	LankMessage string
	LankMessageForeground termbox.Attribute
	RemainingTime float64
}

type Screen struct {
	matrix [][]*screenCell
}

func (screen *Screen) measureRowLength() int {
	return len(screen.matrix)
}

func (screen *Screen) measureColumnLength() int {
	return len(screen.matrix[0])
}

func (screen *Screen) ForEachCells(
	callback func(
		y int,
		x int,
		symbol rune,
		foreground termbox.Attribute,
		background termbox.Attribute)) {
	for y, row := range screen.matrix {
		for x, cell := range row {
			callback(y, x, cell.symbol, cell.foreground, cell.background)
		}
	}
}

func (screen *Screen) Render(props *ScreenProps) {
	rowLength := screen.measureRowLength()
	columnLength := screen.measureColumnLength()

	// Pad elements with blanks.
	// Set borders.
	for y := 0; y < rowLength; y++ {
		for x := 0; x < columnLength; x++ {
			isTopOrBottomEdge := y == 0 || y == rowLength-1
			isLeftOrRightEdge := x == 0 || x == columnLength-1
			symbol := ' '
			switch {
			case isTopOrBottomEdge && isLeftOrRightEdge:
				symbol = '+'
			case isTopOrBottomEdge && !isLeftOrRightEdge:
				symbol = '-'
			case !isTopOrBottomEdge && isLeftOrRightEdge:
				symbol = '|'
			}
			cell := screen.matrix[y][x]
			cell.render(&ScreenCellProps{
				Symbol: symbol,
				Foreground: termbox.ColorWhite,
				Background: termbox.ColorBlack,
			})
		}
	}

	// Place the field.
	fieldPosition := &utils.MatrixPosition{Y: 2, X: 2}
	for y, rowProps := range props.FieldCells {
		for x, cellProps := range rowProps {
			cell := screen.matrix[y + fieldPosition.GetY()][x + fieldPosition.GetX()]
			cell.render(cellProps)
		}
	}

	// Prepare texts.
	texts := make([]*screenText, 0)
	remainingTimeText := fmt.Sprintf("%4.1f", props.RemainingTime)
	timeText := &screenText{
		Position: &utils.MatrixPosition{Y: 3, X: 25},
		Text: fmt.Sprintf("Time : %s", remainingTimeText),
		Foreground: termbox.ColorWhite,
	}
	texts = append(texts, timeText)
	floorNumberText := &screenText{
		Position: &utils.MatrixPosition{Y: 4, X: 25},
		Text: fmt.Sprintf("Floor: %2d", props.FloorNumber),
		Foreground: termbox.ColorWhite,
	}
	texts = append(texts, floorNumberText)
	if props.LankMessage != "" {
		lankText := &screenText{
			Position: &utils.MatrixPosition{Y: 5, X: 27},
			Text: props.LankMessage,
			Foreground: props.LankMessageForeground,
		}
		texts = append(texts, lankText)
	}

	// Place texts.
	for _, textInstance := range texts {
		for deltaX, character := range textInstance.Text {
			cell := screen.matrix[textInstance.Position.GetY()][textInstance.Position.GetX() + deltaX]
			cell.render(&ScreenCellProps{
				Symbol: character,
				Foreground: textInstance.Foreground,
				Background: termbox.ColorBlack,
			})
		}
	}
}

func CreateScreen(rowLength int, columnLength int) *Screen {
	matrix := make([][]*screenCell, rowLength)
	for y := 0; y < rowLength; y++ {
		row := make([]*screenCell, columnLength)
		for x := 0; x < columnLength; x++ {
			cell := &screenCell{}
			cell.render(&ScreenCellProps{
				Symbol:          '_',
				Foreground: termbox.ColorWhite,
				Background: termbox.ColorBlack,
			})
			row[x] = cell
		}
		matrix[y] = row
	}

	return &Screen{
		matrix: matrix,
	}
}
