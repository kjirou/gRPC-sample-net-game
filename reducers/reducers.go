package reducers

import(
	"github.com/pkg/errors"
	"github.com/kjirou/gRPC-sample-net-game/models"
	"github.com/kjirou/gRPC-sample-net-game/utils"
	"time"
)

type FourDirection int
const (
	FourDirectionUp FourDirection = iota
	FourDirectionRight
	FourDirectionDown
	FourDirectionLeft
)

// TODO: 一部の操作は接続中プレイヤーが操作中のAvatorの抽出が前提になるはず。

func proceedMainLoopFrame(state *models.State, elapsedTime time.Duration) (*models.State, error) {
	game := state.GetGame()
	field := state.GetField()

	// Clean field effects.
	field.CleanFieldEffects()

	// Remove timeouted field effects.
	newFieldEffects := make([]*models.FieldEffect, 0)
	for _, fieldEffect := range state.FieldEffects {
		if fieldEffect.CalculateRemainingDuration(state.GetMainLoopNumber()) > 0 {
			newFieldEffects = append(newFieldEffects, fieldEffect)
		}
	}
	state.FieldEffects = newFieldEffects

	// Place field effects to field elements.
	for _, fieldEffect := range state.FieldEffects {
		for _, areaFlagment := range fieldEffect.Area {
			fieldElement, fieldAtOk := field.At(areaFlagment)
			if fieldAtOk {
				fieldElement.FieldEffects = append(fieldElement.FieldEffects, fieldEffect)
			}
		}
	}

	// TODO: Apply field effects to objects.

	// In the game.
	if game.IsStarted() && !game.IsFinished() {
		// Time over of this game.
		remainingTime := game.CalculateRemainingTime(state.GetExecutionTime())
		if remainingTime == 0 {
			game.Finish()
		}
	}

	state.AlterExecutionTime(elapsedTime)
	state.IncrementMainLoopNumber()

	return state, nil
}

func AdvanceOnlyTime(state models.State, elapsedTime time.Duration) (*models.State, error) {
	return proceedMainLoopFrame(&state, elapsedTime)
}

func StartOrRestartGame(state models.State, elapsedTime time.Duration) (*models.State, error) {
	game := state.GetGame()
	//field := state.GetField()

	// Start the new game.
	game.Reset()
	game.Start(state.GetExecutionTime())

	return proceedMainLoopFrame(&state, elapsedTime)
}

func WalkHero(state models.State, elapsedTime time.Duration, direction FourDirection) (*models.State, error) {
	//game := state.GetGame()
	field := state.GetField()

	avatorElement, avatorElementOk := field.FindElementWithAvatorIfPossible()
	if !avatorElementOk {
		return &state, nil
	}

	nextY := avatorElement.GetPosition().GetY()
	nextX := avatorElement.GetPosition().GetX()
	switch direction {
	case FourDirectionUp:
		nextY -= 1
	case FourDirectionRight:
		nextX += 1
	case FourDirectionDown:
		nextY += 1
	case FourDirectionLeft:
		nextX -= 1
	}
	position := avatorElement.GetPosition()
	nextPosition := &utils.MatrixPosition{
		Y: nextY,
		X: nextX,
	}
	nextElement, nextElementOk := field.At(nextPosition)
	if nextElementOk {
		_, nextElementObjectExists := nextElement.GetFieldObjectIfPossible()
		if !nextElementObjectExists {
			moveObjectErr := field.MoveObject(position, nextPosition)
			if moveObjectErr != nil {
				return &state, errors.WithStack(moveObjectErr)
			}
		}
	}
	return proceedMainLoopFrame(&state, elapsedTime)
}

func HeroActs(state models.State, elapsedTime time.Duration) (*models.State, error) {
	field := state.GetField()

	avatorElement, avatorElementOk := field.FindElementWithAvatorIfPossible()
	if !avatorElementOk {
		return &state, nil
	}
	avatorObject, _ := avatorElement.GetFieldObjectIfPossible()

	additionalFieldEffects := avatorObject.Act(state.GetMainLoopNumber(), avatorElement.GetPosition())
	state.FieldEffects = append(state.FieldEffects, additionalFieldEffects...)

	return proceedMainLoopFrame(&state, elapsedTime)
}
