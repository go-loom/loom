assert
======

assertions for tests written in go, junit style.

Example usage:

```
func TestCats(t *testing.T) {
  assert := Assert(t)

  cat1 := &Cat{}
  cat2 := &Cat{}
  // Cat1 and Cat2 start out as identical cats (same position, etc.)
  assert.Equal(cat1, cat2)
  // But they are not the SAME cat
  assert.NotSame(cat1, cat2)
  // However, identity should hold
  assert.Same(cat1, cat1)

  // Cat1 gets bored.
  cat1.Jump(3)
  // Now they are in different positions and therefore not equal
  assert.Equal(cat1.Position, Position{3, 0, 0})
  assert.NotEqual(cat1, cat2)

  // Feed cat 1
  cat1.Feed()
  assert.True(cat1.isFed, "Cat %v not fed", 1)
  assert.False(cat2.isFed, "Cat %v is fed", 2)

  // Cats can only be pet if they have been fed
  err1 := cat1.Pet()
  err2 := cat2.Pet()
  assert.Nil(err1)
  assert.NotNil(err2)
}
```

When an assertion fails, the test stops and an error is printed out
including the line on which it failed and a helpful error message.
For example, if we change the Position assertion to Position{0, 3, 0},
it will print the following:

```
--- FAIL: TestAssert (0.00 seconds)
  assert.go:35:
    assert_test.go:56    <== LINE NUMBER IN YOUR TEST
    Expected: {0 3 0}
    Received: {3 0 0}
FAIL
FAIL  github.com/seanpont/assert  0.008s
```

It's so easy a kitten could do it!

