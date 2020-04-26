package models

import (
	"github.com/kjirou/gRPC-sample-net-game/utils"
)

type FieldObject interface {
	Act(mainLoopNumber int, position *utils.MatrixPosition) []*FieldEffect
	GetClass() string
}

// Avator
type avatorFieldObject struct {
}

func (object *avatorFieldObject) GetClass() string {
	return "avator"
}

func (object *avatorFieldObject) Act(mainLoopNumber int, position *utils.MatrixPosition) []*FieldEffect {
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

// Wall
type wallFieldObject struct {
}

func (object *wallFieldObject) GetClass() string {
	return "wall"
}

func (object *wallFieldObject) Act(mainLoopNumber int, position *utils.MatrixPosition) []*FieldEffect {
	return make([]*FieldEffect, 0)
}
