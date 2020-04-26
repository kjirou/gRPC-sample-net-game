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

type FieldElement struct {
	FieldEffects []*FieldEffect
	fieldObject FieldObject
	floorObjectClass string
	position *utils.MatrixPosition
}

func (fieldElement *FieldElement) GetPosition() *utils.MatrixPosition {
	return fieldElement.position
}

// TODO: 障害物の場合は 0or1 個だが、非障害物の場合は 0-n 個になる。障害物前提にできるかがまだわからない。
func (fieldElement *FieldElement) GetFieldObjectIfPossible() (FieldObject, bool) {
	return fieldElement.fieldObject, fieldElement.fieldObject != nil
}

func (fieldElement *FieldElement) SetFieldObject(fieldObject FieldObject) {
	fieldElement.fieldObject = fieldObject
}

func (fieldElement *FieldElement) GetFloorObjectClass() string {
	return fieldElement.floorObjectClass
}

func (fieldElement *FieldElement) UpdateFloorObjectClass(class string) {
	fieldElement.floorObjectClass = class
}

type Field struct {
	matrix [][]*FieldElement
}

// TODO: 最終的には隠蔽するか、Field.Matrix で露出するかのどちらかにする。
func (field *Field) GetMatrix() [][]*FieldElement {
	return field.matrix
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

func (field *Field) FindElementWithAvatorIfPossible() (*FieldElement, bool) {
	for _, row := range field.matrix {
		for _, element := range row {
			object, objectOk := element.GetFieldObjectIfPossible()
			if objectOk {
				// TODO: 本来は複数のアバターが存在するから、検索条件が不十分なはず。
				if object.GetClass() == "avator" {
					return element, true
				}
			}
		}
	}
	return nil, false
}

func (field *Field) MoveObject(from *utils.MatrixPosition, to *utils.MatrixPosition) error {
	fromElement, fromElementOk := field.At(from)
	if !fromElementOk {
		return errors.New("The `from` position does not exist on the field.")
	} else {
		_, fromElementObjectExists := fromElement.GetFieldObjectIfPossible()
		if !fromElementObjectExists {
			return errors.New("The object to be moved does not exist.")
		}
	}

	toElement, toElementOk := field.At(to)
	if !toElementOk {
		return errors.New("The `to` position does not exist on the field.")
	} else {
		_, toElementObjectExists := toElement.GetFieldObjectIfPossible()
		if toElementObjectExists {
			return errors.New("An object exists at the destination.")
		}
	}

	fromElementObject, _ := fromElement.GetFieldObjectIfPossible()
	fromElement.SetFieldObject(nil)
	toElement.SetFieldObject(fromElementObject)

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
	isFinished bool
	// A snapshot of `state.executionTime` when a game has started.
	startedAt time.Duration
}

func (game *Game) Reset() {
	zeroDuration, _ := time.ParseDuration("0s")
	game.startedAt = zeroDuration
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

func (game *Game) Start(executionTime time.Duration) {
	game.startedAt = executionTime
}

func (game *Game) Finish() {
	game.isFinished = true
}

type indexedFieldObject struct {
	FieldObject FieldObject
	Position *utils.MatrixPosition
}

type State struct {
	FieldEffects []*FieldEffect
	IndexedFieldObjects []*indexedFieldObject
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
	var hero FieldObject = &avatorFieldObject{}
	heroFieldElement.SetFieldObject(hero)

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
				var wall FieldObject = &wallFieldObject{}
				elem.SetFieldObject(wall)
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

func AppendIndexedFieldObject(indexedFieldObjects []*indexedFieldObject, fieldObject FieldObject) error {
	return nil
}

func RemoveIndexedFieldObjectByFieldObject(
	indexedFieldObjects []*indexedFieldObject,
	target FieldObject) ([]*indexedFieldObject, error) {
	newIndexedFieldObjects := make([]*indexedFieldObject, 0)
	for i, indexedFieldObject := range indexedFieldObjects {
		if indexedFieldObject.FieldObject == target {
			if indexedFieldObject.Position != nil {
				return newIndexedFieldObjects, errors.New("The targetted field object is still in placed.")
			} else {
				newIndexedFieldObjects = append(newIndexedFieldObjects, indexedFieldObjects[:i]...)
				newIndexedFieldObjects = append(newIndexedFieldObjects, indexedFieldObjects[i+1:]...)
				return newIndexedFieldObjects, nil
			}
		}
	}
	return newIndexedFieldObjects, errors.New("The targetted field object does not exist.")
}

func PlaceFieldObject(
	field *Field,
	indexedFieldObjects []*indexedFieldObject,
	fieldObject FieldObject,
	to *utils.MatrixPosition) error {
	return nil
}

func UnplaceFieldObject(
	field *Field,
	indexedFieldObjects []*indexedFieldObject,
	fieldObject FieldObject) error {
	return nil
}
