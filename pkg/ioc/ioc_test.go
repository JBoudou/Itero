// Itero - Online iterative vote application
// Copyright (C) 2021 Joseph Boudou
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package ioc

import (
	"errors"
	"fmt"
	"testing"
)

func TestLocator_New(t *testing.T) {
	locator := New()
	if locator == nil {
		t.Errorf("New returned nil")
	}

	var same *Locator
	err := locator.Inject(&same)
	if err != nil {
		t.Errorf("Inject returned error %v.", err)
	}
	if same != locator {
		t.Errorf("Inject returned another locator")
	}
}

func TestLocator_Sub_New(t *testing.T) {
	locator := New()
	sub := locator.Sub()
	if sub == locator {
		t.Errorf("Sub returned the parent")
	}

	var same *Locator
	err := sub.Inject(&same)
	if err != nil {
		t.Errorf("Inject returned error %v.", err)
	}
	if same == locator {
		t.Errorf("Inject returned the parent")
	}
	if same != sub {
		t.Errorf("Inject returned another locator")
	}
}

type testToken struct {
	id  uint32
	grp uint32
}

var nextTextTokenId uint32 = 0

func newPairToken(grp uint32) func() testToken {
	return func() testToken {
		ret := testToken{id: nextTextTokenId, grp: grp}
		nextTextTokenId += 1
		return ret
	}
}

func (self testToken) Is(other testToken) bool {
	return self.id == other.id
}

func TestLocator_Bind(t *testing.T) {
	locator := New()

	err := locator.Bind(newPairToken(1))
	if err != nil {
		t.Errorf("Bind returned %v.", err)
	}

	var first testToken
	err = locator.Inject(&first)
	if err != nil {
		t.Errorf("Inject returned %v.", err)
	}
	if first.grp != 1 {
		t.Errorf("Wrong grp %d.", first.grp)
	}

	var second testToken
	err = locator.Inject(&second)
	if err != nil {
		t.Errorf("Inject returned %v.", err)
	}
	if !first.Is(second) {
		t.Errorf("First and second are not the same")
	}

	err = locator.Bind(func() (testToken, error) { return newPairToken(2)(), nil })
	if err != nil {
		t.Errorf("Bind returned %v.", err)
	}

	var third testToken
	err = locator.Inject(&third)
	if err != nil {
		t.Errorf("Inject returned %v.", err)
	}
	if third.grp != 2 {
		t.Errorf("Wrong grp %d.", third.grp)
	}
	if first.Is(third) {
		t.Errorf("First and third are the same")
	}
}

func TestLocator_Bind_Errors(t *testing.T) {
	tests := []struct {
		name    string
		factory interface{}
		expect  error
	}{
		{
			name:    "Not a function",
			factory: testToken{},
			expect:  NotFactory,
		},
		{
			name:    "No returned value",
			factory: func() {},
			expect:  NotFactory,
		},
		{
			name:    "Two returned values",
			factory: func() (uint8, uint16) { return 1, 2 },
			expect:  NotFactory,
		},
		{
			name:    "Three returned values",
			factory: func() (uint8, error, uint32) { return 1, nil, 4 },
			expect:  NotFactory,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			locator := New()
			err := locator.Bind(tt.factory)
			if err == nil {
				t.Error("No error.")
			}
			if !errors.Is(err, tt.expect) {
				t.Errorf("Wrong error. Got %v. Expect %v.", err, tt.expect)
			}
		})
	}
}

func TestLocator_Bind_Sub(t *testing.T) {
	parent := New()
	parent.Bind(newPairToken(1))

	child := parent.Sub()
	err := child.Bind(newPairToken(2))
	if err != nil {
		t.Errorf("Bind error %v.", err)
	}

	getBoth := func(parentValue interface{}, childValue interface{}) {
		err = parent.Inject(parentValue)
		if err != nil {
			t.Errorf("Parent Inject error %v.", err)
		}
		err = child.Inject(childValue)
		if err != nil {
			t.Errorf("Child Inject error %v.", err)
		}
	}

	var parentToken, childToken testToken
	getBoth(&parentToken, &childToken)

	if parentToken.Is(childToken) {
		t.Errorf("Parent and child tokens are the same.")
	}
	if childToken.grp != 2 {
		t.Errorf("Wrong grp %d for child token.", childToken.grp)
	}

	parent.Bind(func() int { return 27 })
	var parentInt, childInt int
	getBoth(&parentInt, &childInt)
	if parentInt != childInt {
		t.Errorf("Parent and child integer are not the same.")
	}
}

func TestLocator_Inject_One(t *testing.T) {
	locator := New()
	locator.Bind(newPairToken(1))

	var direct testToken
	err := locator.Inject(&direct)
	if err != nil {
		t.Errorf("Inject value returned %v.", err)
	}

	var byfun testToken
	err = locator.Inject(func(tok testToken) { byfun = tok })
	if err != nil {
		t.Errorf("Inject function returned %v.", err)
	}
	if !direct.Is(byfun) {
		t.Errorf("Direct and by-function are not the same.")
	}

	err = locator.Inject(func(loc *Locator, tok testToken) {
		if loc != locator {
			t.Errorf("Wrong locator.")
		}
		byfun = tok
	})
	if !direct.Is(byfun) {
		t.Errorf("Direct and by-function are not the same.")
	}
}

func TestLocator_Inject_Two(t *testing.T) {
	locator := New()
	locator.Bind(newPairToken(1))

	var result uint32
	err := locator.Inject(func(loc *Locator, tok testToken) uint32 {
		if loc != locator {
			t.Errorf("Wrong locator")
		}
		return tok.grp
	}, &result)
	if err != nil {
		t.Errorf("Inject error %v.", err)
	}
	if result != 1 {
		t.Errorf("Wrong result value. Got %d. Expect 1.", result)
	}

	var generic interface{ Is(testToken) bool }
	err = locator.Inject(func(tok testToken) testToken { return tok }, &generic)
	if err != nil {
		t.Errorf("Inject error %v.", err)
	}
}

func TestLocator_Inject_Int(t *testing.T) {
	tests := []struct {
		name   string
		chain  []interface{}
		expect int
	}{
		{
			name:   "Empty function",
			chain:  []interface{}{func() {}},
			expect: 0,
		},
		{
			name: "Documentation example",
			chain: []interface{}{
				func() int { return 1 },
				func() int { return 2 },
				func() interface{} { return 3 },
			},
			expect: 2,
		},
		{
			name: "Chain",
			chain: []interface{}{
				func(a int) int { return a + 2 },
				func(b int) int { return b * 3 },
				func(c int) int { return c - 1 },
			},
			expect: 5,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			locator := New()
			locator.Bind(func() int { return 0 })

			var got int
			tt.chain = append(tt.chain, &got)
			locator.Inject(tt.chain...)

			if got != tt.expect {
				t.Errorf("Got %d. Expect %d.", got, tt.expect)
			}
		})
	}
}

func TestLocator_Inject_Error(t *testing.T) {
	var vInt int
	var vToken testToken
	customError := errors.New("Custom")

	tests := []struct {
		name     string
		factory  interface{}
		receptor interface{}
		expect   error
	}{
		{
			name:     "Not a function",
			factory:  testToken{},
			receptor: &vToken,
			expect:   NotReceptor,
		},
		{
			name:    "Nil receptor",
			factory: func() {},
			expect:  NotReceptor,
		},
		{
			name:     "Not a pointer",
			factory:  func() int { return 42 },
			receptor: vInt,
			expect:   NotReceptor,
		},
		{
			name:     "Not found",
			factory:  func(z uint32) int { return int(z) },
			receptor: &vInt,
			expect:   NotFound,
		},
		{
			name:     "Error from factory",
			factory:  func() (int, error) { return 0, customError },
			receptor: &vInt,
			expect:   customError,
		},
		{
			name:     "Assignable",
			factory:  func(x interface{ Is(testToken) bool }) {},
			receptor: &vInt,
			expect:   NotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			locator := New()
			locator.Bind(newPairToken(1))
			locator.Bind(func() int { return 27 })

			err := locator.Inject(tt.factory, tt.receptor)
			if !errors.Is(err, tt.expect) {
				t.Errorf("Wrong error. Got %v. Expect %v.", err, tt.expect)
			}
		})
	}
}

func TestLocator_Fresh(t *testing.T) {
	locator := New()
	locator.Bind(newPairToken(1))

	var first testToken
	err := locator.Fresh(&first)
	if err != nil {
		t.Errorf("Fresh returned %v.", err)
	}

	var singleton testToken
	err = locator.Inject(&singleton)
	if err != nil {
		t.Errorf("Inject returned %v.", err)
	}
	if singleton.Is(first) {
		t.Errorf("Singleton is the same as the first fresh.")
	}

	var second testToken
	err = locator.Fresh(&second)
	if err != nil {
		t.Errorf("Fresh returned %v.", err)
	}
	if singleton.Is(first) {
		t.Errorf("Singleton is the same as the second fresh.")
	}
	if singleton.Is(first) {
		t.Errorf("The second fresh is the same as the first fresh.")
	}

	var sub *Locator
	err = locator.Fresh(&sub)
	if err != nil {
		t.Errorf("Fresh returned %v.", err)
	}
	if sub == locator {
		t.Errorf("Sub is same as parent")
	}
}

func TestLocator_Fresh_Error(t *testing.T) {
	var vInt int
	customError := errors.New("Custom")

	tests := []struct {
		name     string
		factory  interface{}
		receptor interface{}
		expect   error
	}{
		{
			name:     "NotFound value",
			receptor: &vInt,
			expect:   NotFound,
		},
		{
			name:     "NotReceptor value",
			receptor: 42,
			expect:   NotReceptor,
		},
		{
			name:     "NotReceptor function",
			receptor: func(tok testToken) {},
			expect:   NotReceptor,
		},
		{
			name:     "Error from factory",
			factory:  func() (int, error) { return 0, customError },
			receptor: &vInt,
			expect:   customError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			locator := New()
			locator.Bind(newPairToken(1))

			if tt.factory != nil {
				err := locator.Bind(tt.factory)
				if err != nil {
					t.Errorf("Error in set: %v.", err)
				}
			}

			err := locator.Fresh(tt.receptor)
			if !errors.Is(err, tt.expect) {
				t.Errorf("Wrong error. Got %v. Expect %v.", err, tt.expect)
			}
		})
	}
}

func TestLocator_Refresh(t *testing.T) {
	locator := New()
	locator.Bind(newPairToken(1))

	var first testToken
	err := locator.Inject(&first)
	if err != nil {
		t.Errorf("Inject returned %v.", err)
	}

	var fresh testToken
	err = locator.Refresh(&fresh)
	if err != nil {
		t.Errorf("Refresh returned %v.", err)
	}
	if fresh.Is(first) {
		t.Errorf("Refresh value is not fresh.")
	}

	var second testToken
	err = locator.Inject(&second)
	if err != nil {
		t.Errorf("Inject returned %v.", err)
	}
	if second.Is(first) {
		t.Errorf("Refresh did not change the singleton value.")
	}
	if !second.Is(fresh) {
		t.Errorf("Singleton value is not the same as the fresh value.")
	}
}

func TestLocator_Refresh_Sub(t *testing.T) {
	parent := New()
	parent.Bind(newPairToken(1))
	child := parent.Sub()

	var fresh testToken
	err := child.Refresh(&fresh)
	if err != nil {
		t.Errorf("Refresh returned %v.", err)
	}

	var first testToken
	err = parent.Inject(&first)
	if err != nil {
		t.Errorf("Inject returned %v.", err)
	}
	var second testToken
	err = child.Inject(&second)
	if err != nil {
		t.Errorf("Inject returned %v.", err)
	}
	if second.Is(first) {
		t.Errorf("Refresh did not change the singleton value.")
	}
	if !second.Is(fresh) {
		t.Errorf("Singleton value is not the same as the fresh value.")
	}
}

func TestLocator_Refresh_Error(t *testing.T) {
	var vInt int
	customError := errors.New("Custom")

	tests := []struct {
		name     string
		factory  interface{}
		receptor interface{}
		expect   error
	}{
		{
			name:     "NotFound value",
			receptor: &vInt,
			expect:   NotFound,
		},
		{
			name:     "NotReceptor value",
			receptor: 42,
			expect:   NotReceptor,
		},
		{
			name:     "NotReceptor function",
			receptor: func(tok testToken) {},
			expect:   NotReceptor,
		},
		{
			name:     "Error from factory",
			factory:  func() (int, error) { return 0, customError },
			receptor: &vInt,
			expect:   customError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			locator := New()
			locator.Bind(newPairToken(1))

			if tt.factory != nil {
				err := locator.Bind(tt.factory)
				if err != nil {
					t.Errorf("Error in set: %v.", err)
				}
			}

			err := locator.Refresh(tt.receptor)
			if !errors.Is(err, tt.expect) {
				t.Errorf("Wrong error. Got %v. Expect %v.", err, tt.expect)
			}
		})
	}
}

// Examples

func ExampleLocator_Inject() {
	loc := New()
	var v int
	loc.Inject(func() int { return 1 }, func() int { return 2 }, func() interface{} { return 3 }, &v)
	fmt.Println(v)
	// Output:
	// 2
}
