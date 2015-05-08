package assert

import (
	"errors"
	"math"
	"testing"
)

type Position struct {
	X, Y, Theta float64
}

func (p Position) Move(distance float64) Position {
	return Position{
		X:     p.X + math.Cos(p.Theta)*distance,
		Y:     p.Y + math.Sin(p.Theta)*distance,
		Theta: p.Theta,
	}
}

type Cat struct {
	Position Position
	isFed    bool
}

func (cat *Cat) Jump(distance float64) {
	cat.Position = cat.Position.Move(distance)
}

func (cat *Cat) Feed() {
	cat.isFed = true
}

func (cat *Cat) Pet() error {
	if !cat.isFed {
		return errors.New("CLAW!")
	}
	return nil
}

func TestAssert(t *testing.T) {
	assert := Assert(t)

	cat1 := &Cat{}
	cat2 := &Cat{}
	// Cat1 and Cat2 start out as identical cats (same position, etc.)
	assert.Equal(cat1, cat2)
	// But they are not the SAME cat
	assert.NotSame(cat1, cat2)
	// The identity should hold
	assert.Same(cat1, cat1)

	// Cat1 gets bored.
	cat1.Jump(3)
	// Now they are in different positions, and therefore not equal
	assert.Equal(cat1.Position, Position{3, 0, 0})
	assert.NotEqual(cat1, cat2)

	// Cats can only be pet after they have been fed
	cat1.Feed()
	assert.True(cat1.isFed, "Cat %v not fed", 1)
	assert.False(cat2.isFed, "Cat %v is fed", 2)

	// Cats can only be pet if they have been fed
	err1 := cat1.Pet()
	err2 := cat2.Pet()
	assert.Nil(err1)
	assert.NotNil(err2)
}
