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
	field := state.GetField()

	// Generate a new maze.
	// Remove the hero.
	err := field.ResetMaze()
	if err != nil {
		return &state, errors.WithStack(err)
	}

	// Replace the hero.
	heroFieldElement, _ := field.At(models.HeroPosition)
	heroFieldElement.UpdateObjectClass("hero")

	// Start the new game.
	game.Reset()
	game.Start(state.GetExecutionTime())

	return proceedMainLoopFrame(&state, elapsedTime)
}

func WalkHero(state models.State, elapsedTime time.Duration, direction FourDirection) (*models.State, error) {
	game := state.GetGame()
	if game.IsFinished() {
		return &state, nil
	}

	field := state.GetField()
	element, getElementOfHeroErr := field.GetElementOfHero()
	if getElementOfHeroErr != nil {
		return &state, errors.WithStack(getElementOfHeroErr)
	}
	nextY := element.GetPosition().GetY()
	nextX := element.GetPosition().GetX()
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
	position := element.GetPosition()
	nextPosition := &utils.MatrixPosition{
		Y: nextY,
		X: nextX,
	}
	if nextPosition.Validate(field.MeasureRowLength(), field.MeasureColumnLength()) {
		element, elementOk := field.At(nextPosition)
		if !elementOk {
			return &state, errors.Errorf("The %v position does not exist on the field.", nextPosition)
		} else if element.IsObjectEmpty() {
			err := field.MoveObject(position, nextPosition)
			return &state, errors.WithStack(err)
		}
	}
	return proceedMainLoopFrame(&state, elapsedTime)
}

func HeroActs(state models.State, elapsedTime time.Duration) (*models.State, error) {
	field := state.GetField()

	heroElement, getElementOfHeroErr := field.GetElementOfHero()
	if getElementOfHeroErr != nil {
		return &state, errors.WithStack(getElementOfHeroErr)
	}

	adventurer := models.Adventurer{}
	additionalFieldEffects := adventurer.Act(state.GetMainLoopNumber(), heroElement.GetPosition())
	state.FieldEffects = append(state.FieldEffects, additionalFieldEffects...)
	return proceedMainLoopFrame(&state, elapsedTime)
}
