package models

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kjirou/gRPC-sample-net-game/utils"
)

func TestField_At_NotTD(t *testing.T) {
	field := createField(2, 3)

	t.Run("指定した位置の要素を取得できる", func(t *testing.T) {
		element, _ := field.At(&utils.MatrixPosition{Y: 1, X: 2})
		if element.GetPosition().GetY() != 1 {
			t.Fatal("Y が違う")
		} else if element.GetPosition().GetX() != 2 {
			t.Fatal("X が違う")
		}
	})

	t.Run("存在しない位置を指定したとき", func(t *testing.T) {
		type testCase struct {
			Y int
			X int
		}
		var testCases []testCase
		testCases = append(testCases, testCase{Y: -1, X: 0})
		testCases = append(testCases, testCase{Y: 2, X: 0})
		testCases = append(testCases, testCase{Y: 0, X: -1})
		testCases = append(testCases, testCase{Y: 0, X: 3})
		for _, tc := range testCases {
			tc := tc
			t.Run(fmt.Sprintf("Y=%d,X=%dの第2戻り値はfalseを返す", tc.Y, tc.X), func(t *testing.T) {
				_, ok := field.At(&utils.MatrixPosition{Y: tc.Y, X: tc.X})
				if ok {
					t.Fatal("falseを返さない")
				}
			})
		}
	})
}

func TestGame_CalculateRemainingTime_NotTD(t *testing.T) {
	game := &Game{}

	t.Run("リセット直後は30を返す", func(t *testing.T) {
		game.Reset()
		executionTime, _ := time.ParseDuration("2s")
		remainingTime := game.CalculateRemainingTime(executionTime)
		if remainingTime.Seconds() != 30 {
			t.Fatal("30ではない")
		}
	})

	t.Run("最小で0を返す", func(t *testing.T) {
		game.Reset()
		startTime, _ := time.ParseDuration("1s")
		game.Start(startTime)
		currentTime, _ := time.ParseDuration("999s")
		remainingTime := game.CalculateRemainingTime(currentTime)
		if remainingTime.Seconds() != 0 {
			t.Fatal("0ではない")
		}
	})
}

func TestRemoveIndexedFieldObjectByFieldObject_NotTD(t *testing.T) {
	t.Run("目標が存在して未配置のとき、リストからそれを削除できる", func(t *testing.T) {
		target := &wallFieldObject{}
		one := &indexedFieldObject{
			FieldObject: &wallFieldObject{},
		}
		two := &indexedFieldObject{
			FieldObject: target,
		}
		three := &indexedFieldObject{
			FieldObject: &wallFieldObject{},
		}
		indexedFieldObjects := []*indexedFieldObject{
			one,
			two,
			three,
		}
		newIndexedFieldObjects, err := RemoveIndexedFieldObjectByFieldObject(indexedFieldObjects, target)
		if err != nil {
			t.Fatal("エラーを返している")
		}
		if !reflect.DeepEqual(
			newIndexedFieldObjects,
			[]*indexedFieldObject{
				one,
				three,
			},
		) {
			t.Fatal("削除されていないか誤った要素を削除している")
		}
	})

	t.Run("目標が存在して配置中のとき、エラーを返す", func(t *testing.T) {
		target := &wallFieldObject{}
		indexedFieldObjects := []*indexedFieldObject{
			&indexedFieldObject{
				FieldObject: target,
				Position: &utils.MatrixPosition{},
			},
		}
		_, err := RemoveIndexedFieldObjectByFieldObject(indexedFieldObjects, target)
		if err == nil {
			t.Fatal("エラーを返さない")
		} else if !strings.Contains(err.Error(), " in placed") {
			t.Fatal("意図したエラーメッセージではない")
		}
	})

	t.Run("目標が存在しないとき、エラーを返す", func(t *testing.T) {
		target := &wallFieldObject{}
		indexedFieldObjects := []*indexedFieldObject{
			&indexedFieldObject{
				FieldObject: &wallFieldObject{},
				Position: &utils.MatrixPosition{},
			},
		}
		_, err := RemoveIndexedFieldObjectByFieldObject(indexedFieldObjects, target)
		if err == nil {
			t.Fatal("エラーを返さない")
		} else if !strings.Contains(err.Error(), " does not exist") {
			t.Fatal("意図したエラーメッセージではない")
		}
	})
}
