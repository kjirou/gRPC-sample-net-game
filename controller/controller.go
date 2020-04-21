package controller

//
// NOTE: 全体的な設計について。
//
// Inputs   = 経過時間とキー入力が本アプリケーションが認識する外部入力である。
//   |
// Reducers = いち Reducer 関数は、その関数が対応する Input と現在の Models を組み合わせて、
//   |          次の Models を生成して返す。
// Models   = 本アプリケーションの正規化された状態である。
//   |        これ以下の状態は全て Models の写像として生成される。
// Props    = Views 側が要求する Views への更新クエリである。
//   |        Models の写像として生成される。
// Views    = 端末出力を抽象化した層である。
//   |        termbox へ渡すためのセルの矩形の集合を生成することが目的である。
// Outputs  = 基本的には termbox 経由で端末へ出力する。
//            デバッグ用に標準出力をすることもある。
//

import (
	"github.com/kjirou/gRPC-sample-net-game/models"
	"github.com/kjirou/gRPC-sample-net-game/utils"
	"github.com/kjirou/gRPC-sample-net-game/reducers"
	"github.com/kjirou/gRPC-sample-net-game/views"
	"github.com/nsf/termbox-go"
	"github.com/pkg/errors"
	"math"
	"time"
)

func mapFieldElementToScreenCellProps(fieldElement *models.FieldElement) *views.ScreenCellProps {
	symbol := '.'
	fg := termbox.ColorWhite
	bg := termbox.ColorBlack
	if !fieldElement.IsObjectEmpty() {
		switch fieldElement.GetObjectClass() {
		case "hero":
			symbol = '@'
			fg = termbox.ColorMagenta
		case "wall":
			symbol = '#'
			fg = termbox.ColorYellow
		default:
			symbol = '?'
		}
	} else {
		switch fieldElement.GetFloorObjectClass() {
		case "upstairs":
			symbol = '<'
			fg = termbox.ColorGreen
		}
	}
	return &views.ScreenCellProps{
		Symbol: symbol,
		Foreground: fg,
		Background: bg,
	}
}

func mapStateModelToScreenProps(state *models.State) (*views.ScreenProps, error) {
	game := state.GetGame()
	field := state.GetField()

	heroElement, heroElementErr := field.GetElementOfHero()
	if heroElementErr != nil {
		return nil, errors.WithStack(heroElementErr)
	}
	heroPosition := heroElement.GetPosition()

	// Cells of the field.
	fieldCellsRowLength := 13
	fieldCellsColumnLength := 21
	fieldCellsCenterPosition := &utils.MatrixPosition{
		Y: int(math.Ceil(float64(fieldCellsRowLength)/2-1)),
		X: int(math.Ceil(float64(fieldCellsColumnLength)/2-1)),
	}
	fieldCells := make([][]*views.ScreenCellProps, fieldCellsRowLength)
	for y := 0; y < fieldCellsRowLength; y++ {
		cellsRow := make([]*views.ScreenCellProps, fieldCellsColumnLength)
		for x := 0; x < fieldCellsColumnLength; x++ {
			fieldElement, fieldElementOk := field.At(&utils.MatrixPosition{
				Y: y - fieldCellsCenterPosition.GetY() + heroPosition.GetY(),
				X: x - fieldCellsCenterPosition.GetX() + heroPosition.GetX(),
			})
			if fieldElementOk {
				cellsRow[x] = mapFieldElementToScreenCellProps(fieldElement)
			} else {
				cellsRow[x] = &views.ScreenCellProps{
					Symbol: ' ',
					Foreground: termbox.ColorWhite,
					Background: termbox.ColorBlack,
				}
			}
		}
		fieldCells[y] = cellsRow
	}

	// Lank message.
	lankMessage := ""
	lankMessageForeground := termbox.ColorWhite
	if game.IsFinished() {
		score := game.GetFloorNumber()
		switch {
			case score == 3:
				lankMessage = "Good!"
				lankMessageForeground = termbox.ColorGreen
			case score == 4:
				lankMessage = "Excellent!"
				lankMessageForeground = termbox.ColorGreen
			case score == 5:
				lankMessage = "Marvelous!"
				lankMessageForeground = termbox.ColorGreen
			case score >= 6:
				lankMessage = "Gopher!!"
				lankMessageForeground = termbox.ColorCyan
			default:
				lankMessage = "No good..."
		}
	}

	return &views.ScreenProps{
		FieldCells: fieldCells,
		RemainingTime: game.CalculateRemainingTime(state.GetExecutionTime()).Seconds(),
		FloorNumber: game.GetFloorNumber(),
		LankMessage: lankMessage,
		LankMessageForeground: lankMessageForeground,
	}, nil
}

type Controller struct {
	inputtedCharacter rune
	inputtedKey termbox.Key
	lastMainLoopRanAt time.Time
	state  *models.State
	screen *views.Screen
}

func (controller *Controller) GetScreen() *views.Screen {
	return controller.screen
}

func (controller *Controller) setKeyInputs(ch rune, key termbox.Key) {
	controller.inputtedCharacter = ch
	controller.inputtedKey = key
}

func (controller *Controller) resetKeyInputs() {
	controller.setKeyInputs(0, 0)
}

func (controller *Controller) CalculateIntervalToNextMainLoop(now time.Time) time.Duration {
	// About 60fps.
	intervalOfPurpose := time.Microsecond*16666
	minInterval := time.Microsecond*8333
	nextInterval := intervalOfPurpose
	if !controller.lastMainLoopRanAt.IsZero() {
		actualInterval := now.Sub(controller.lastMainLoopRanAt)
		nextInterval = intervalOfPurpose - (actualInterval - intervalOfPurpose)
		if nextInterval < minInterval {
			nextInterval = minInterval
		} else if nextInterval > intervalOfPurpose {
			nextInterval = intervalOfPurpose
		}
	}
	controller.lastMainLoopRanAt = now
	return nextInterval
}

func (controller *Controller) Dispatch(newState *models.State) error {
	controller.state = newState
	screenProps, err := mapStateModelToScreenProps(controller.state)
	controller.screen.Render(screenProps)
	return err
}

func (controller *Controller) HandleMainLoop(elapsedTime time.Duration) (*models.State, error) {
	ch := controller.inputtedCharacter
	key := controller.inputtedKey
	controller.resetKeyInputs()

	var newState *models.State
	var err error

	switch {
	// Start or restart a game.
	case ch == 's':
		newState, err = reducers.StartOrRestartGame(*controller.state, elapsedTime)
	// Move the hero.
	case key == termbox.KeyArrowUp || ch == 'k':
		newState, err = reducers.WalkHero(*controller.state, elapsedTime, reducers.FourDirectionUp)
	case key == termbox.KeyArrowRight || ch == 'l':
		newState, err = reducers.WalkHero(*controller.state, elapsedTime, reducers.FourDirectionRight)
	case key == termbox.KeyArrowDown || ch == 'j':
		newState, err = reducers.WalkHero(*controller.state, elapsedTime, reducers.FourDirectionDown)
	case key == termbox.KeyArrowLeft || ch == 'h':
		newState, err = reducers.WalkHero(*controller.state, elapsedTime, reducers.FourDirectionLeft)
	default:
		newState, err = reducers.AdvanceOnlyTime(*controller.state, elapsedTime)
	}

	return newState, err
}

func (controller *Controller) HandleKeyPress(ch rune, key termbox.Key) {
	controller.setKeyInputs(ch, key)
}

func CreateController() (*Controller, error) {
	controller := &Controller{}

	state := models.CreateState()
	setWelcomeDataErr := state.SetWelcomeData()
	if setWelcomeDataErr != nil {
		return nil, errors.WithStack(setWelcomeDataErr)
	}

	screen := views.CreateScreen(24, 80)

	controller.resetKeyInputs()
	controller.state = state
	controller.screen = screen
	dispatchErr := controller.Dispatch(state)
	if dispatchErr != nil {
		return nil, errors.WithStack(dispatchErr)
	}

	return controller, nil
}
