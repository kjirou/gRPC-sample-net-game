package models

import (
	"github.com/kjirou/gRPC-sample-net-game/utils"
	"github.com/pkg/errors"
	"time"
)

var HeroPosition = &utils.MatrixPosition{Y: 1, X: 1}
var UpstairsPosition = &utils.MatrixPosition{Y: 11, X: 19}

type FieldEffect struct {
	Area []*utils.MatrixPosition
	// A State.mainLoopNumber when it was created.
	createdAt int
	duration int
}

func (fieldEffect *FieldEffect) CalculateRemainingDuration(mainLoopNumber int) int {
	remainingDuration := fieldEffect.createdAt + fieldEffect.duration - mainLoopNumber
	if remainingDuration < 0 {
		return 0
	}
	return remainingDuration
}

type Adventurer struct {
}

func (adventurer *Adventurer) Act(mainLoopNumber int, position *utils.MatrixPosition) []*FieldEffect {
	fieldEffects := make([]*FieldEffect, 0)
	fieldEffects = append(fieldEffects, &FieldEffect{
		Area: []*utils.MatrixPosition{
			&utils.MatrixPosition{
				Y: position.GetY() - 1,
				X: position.GetX(),
			},
		},
		duration: 3,
		createdAt: mainLoopNumber,
	})
	return fieldEffects
}

type FieldElement struct {
	FieldEffects []*FieldEffect
	floorObjectClass string
	objectClass string
	position *utils.MatrixPosition
}

func (fieldElement *FieldElement) GetPosition() *utils.MatrixPosition {
	return fieldElement.position
}

func (fieldElement *FieldElement) GetObjectClass() string {
	return fieldElement.objectClass
}

func (fieldElement *FieldElement) GetFloorObjectClass() string {
	return fieldElement.floorObjectClass
}

func (fieldElement *FieldElement) IsObjectEmpty() bool {
	return fieldElement.objectClass == "empty"
}

func (fieldElement *FieldElement) UpdateObjectClass(class string) {
	fieldElement.objectClass = class
}

func (fieldElement *FieldElement) UpdateFloorObjectClass(class string) {
	fieldElement.floorObjectClass = class
}

type Field struct {
	matrix [][]*FieldElement
}

func (field *Field) CleanFieldEffects() {
	for _, row := range field.matrix {
		for _, element := range row {
			element.FieldEffects = make([]*FieldEffect, 0)
		}
	}
}

func (field *Field) MeasureRowLength() int {
	return len(field.matrix)
}

func (field *Field) MeasureColumnLength() int {
	return len(field.matrix[0])
}

func (field *Field) At(position *utils.MatrixPosition) (*FieldElement, bool) {
	y := position.GetY()
	x := position.GetX()
	if y < 0 || y > field.MeasureRowLength()-1 {
		return &FieldElement{}, false
	} else if x < 0 || x > field.MeasureColumnLength()-1 {
		return &FieldElement{}, false
	}
	return field.matrix[y][x], true
}

func (field *Field) findElementsByObjectClass(objectClass string) []*FieldElement {
	elements := make([]*FieldElement, 0)
	for _, row := range field.matrix {
		for _, element := range row {
			if element.objectClass == objectClass {
				element_ := element
				elements = append(elements, element_)
			}
		}
	}
	return elements
}

func (field *Field) GetElementOfHero() (*FieldElement, error) {
	elements := field.findElementsByObjectClass("hero")
	if len(elements) == 0 {
		return &FieldElement{}, errors.Errorf("The hero does not exist.")
	} else if len(elements) > 1 {
		return &FieldElement{}, errors.Errorf("There are multiple heroes.")
	}
	return elements[0], nil
}

func (field *Field) MoveObject(from *utils.MatrixPosition, to *utils.MatrixPosition) error {
	fromElement, fromElementOk := field.At(from)
	if !fromElementOk {
		return errors.New("The `from` position does not exist on the field.")
	} else if fromElement.IsObjectEmpty() {
		return errors.New("The object to be moved does not exist.")
	}
	toElement, toElementOk := field.At(to)
	if !toElementOk {
		return errors.New("The `to` position does not exist on the field.")
	} else if !toElement.IsObjectEmpty() {
		return errors.New("An object exists at the destination.")
	}
	toElement.UpdateObjectClass(fromElement.GetObjectClass())
	fromElement.UpdateObjectClass("empty")
	return nil
}

func (field *Field) ResetMaze() error {
	rowLength := field.MeasureRowLength()
	columnLength := field.MeasureColumnLength()
	mazeCells, err := utils.GenerateMaze(rowLength, columnLength)
	if err != nil {
		return err
	}
	for y, mazeRow := range mazeCells {
		for x, mazeCell := range mazeRow {
			element, _ := field.At(&utils.MatrixPosition{Y: y, X: x})
			switch mazeCell.Content {
			case utils.MazeCellContentEmpty:
				element.UpdateObjectClass("empty")
			case utils.MazeCellContentUnbreakableWall:
				element.UpdateObjectClass("wall")
			}
		}
	}
	return nil
}

func createField(y int, x int) *Field {
	matrix := make([][]*FieldElement, y)
	for rowIndex := 0; rowIndex < y; rowIndex++ {
		row := make([]*FieldElement, x)
		for columnIndex := 0; columnIndex < x; columnIndex++ {
			row[columnIndex] = &FieldElement{
				position: &utils.MatrixPosition{
					Y: rowIndex,
					X: columnIndex,
				},
				objectClass: "empty",
				floorObjectClass: "empty",
			}
		}
		matrix[rowIndex] = row
	}
	return &Field{
		matrix: matrix,
	}
}

type Game struct {
	floorNumber int
	isFinished bool
	// A snapshot of `state.executionTime` when a game has started.
	startedAt time.Duration
}

func (game *Game) Reset() {
	zeroDuration, _ := time.ParseDuration("0s")
	game.startedAt = zeroDuration
	game.floorNumber = 1
	game.isFinished = false
}

func (game *Game) IsStarted() bool {
	zeroDuration, _ := time.ParseDuration("0s")
	return game.startedAt != zeroDuration
}

func (game *Game) IsFinished() bool {
	return game.isFinished
}

func (game *Game) CalculateRemainingTime(executionTime time.Duration) time.Duration {
	oneGameTime, _ := time.ParseDuration("30s")
	if game.IsStarted() {
		playtime := executionTime - game.startedAt
		remainingTime := oneGameTime - playtime
		if remainingTime < 0 {
			zeroTime, _ := time.ParseDuration("0s")
			return zeroTime
		}
		return remainingTime
	}
	return oneGameTime
}

func (game *Game) GetFloorNumber() int{
	return game.floorNumber
}

func (game *Game) Start(executionTime time.Duration) {
	game.startedAt = executionTime
}

func (game *Game) Finish() {
	game.isFinished = true
}

type State struct {
	FieldEffects []*FieldEffect
	// This is the total of main loop intervals.
	// It is different from the real time.
	executionTime time.Duration
	field *Field
	game *Game
	mainLoopNumber int
}

func (state *State) GetMainLoopNumber() int {
	return state.mainLoopNumber
}

func (state *State) IncrementMainLoopNumber() {
	state.mainLoopNumber += 1
}

func (state *State) GetExecutionTime() time.Duration {
	return state.executionTime
}

func (state *State) AlterExecutionTime(delta time.Duration) {
	state.executionTime = state.executionTime + delta
}

func (state *State) GetField() *Field {
	return state.field
}

func (state *State) GetGame() *Game {
	return state.game
}

func (state *State) SetWelcomeData() error {
	field := state.GetField()

	// Place a hero to be the player's alter ego.
	heroFieldElement, heroFieldElementOk := field.At(HeroPosition)
	if !heroFieldElementOk {
		return errors.New("The hero's position does not exist on the field.")
	}
	heroFieldElement.UpdateObjectClass("hero")

	// Place an upstairs.
	upstairsFieldElement, upstairsFieldElementOk := field.At(UpstairsPosition)
	if !upstairsFieldElementOk {
		return errors.New("The upstairs' position does not exist on the field.")
	}
	upstairsFieldElement.UpdateFloorObjectClass("upstairs")

	// Place defalt walls.
	fieldRowLength := field.MeasureRowLength()
	fieldColumnLength := field.MeasureColumnLength()
	for y := 0; y < fieldRowLength; y++ {
		for x := 0; x < fieldColumnLength; x++ {
			isTopOrBottomEdge := y == 0 || y == fieldRowLength-1
			isLeftOrRightEdge := x == 0 || x == fieldColumnLength-1
			if isTopOrBottomEdge || isLeftOrRightEdge {
				elem, _ := field.At(&utils.MatrixPosition{Y: y, X: x})
				elem.UpdateObjectClass("wall")
			}
		}
	}

	return nil
}

func CreateState() *State {
	executionTime, _ := time.ParseDuration("0")
	state := &State{
		mainLoopNumber: 1,
		executionTime: executionTime,
		field: createField(13, 21),
		game: &Game{},
	}
	state.game.Reset()
	return state
}
