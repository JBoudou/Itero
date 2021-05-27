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
	"testing"
)

func TestLocator_New(t *testing.T) {
	locator := New()
	if locator == nil {
		t.Errorf("New returned nil")
	}

	var same *Locator
	err := locator.Get(&same)
	if err != nil {
		t.Errorf("Get returned error %v.", err)
	}
	if same != locator {
		t.Errorf("Get returned another locator")
	}
}

func TestLocator_Sub_New(t *testing.T) {
	locator := New()
	sub := locator.Sub()
	if sub == locator {
		t.Errorf("Sub returned the parent")
	}

	var same *Locator
	err := sub.Get(&same)
	if err != nil {
		t.Errorf("Get returned error %v.", err)
	}
	if same == locator {
		t.Errorf("Get returned the parent")
	}
	if same != sub {
		t.Errorf("Get returned another locator")
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

func TestLocator_Set(t *testing.T) {
	locator := New()

	err := locator.Set(newPairToken(1))
	if err != nil {
		t.Errorf("Set returned %v.", err)
	}

	var first testToken
	err = locator.Get(&first)
	if err != nil {
		t.Errorf("Get returned %v.", err)
	}
	if first.grp != 1 {
		t.Errorf("Wrong grp %d.", first.grp)
	}

	var second testToken
	err = locator.Get(&second)
	if err != nil {
		t.Errorf("Get returned %v.", err)
	}
	if !first.Is(second) {
		t.Errorf("First and second are not the same")
	}

	err = locator.Set(func() (testToken, error) { return newPairToken(2)(), nil })
	if err != nil {
		t.Errorf("Set returned %v.", err)
	}

	var third testToken
	err = locator.Get(&third)
	if err != nil {
		t.Errorf("Get returned %v.", err)
	}
	if third.grp != 2 {
		t.Errorf("Wrong grp %d.", third.grp)
	}
	if first.Is(third) {
		t.Errorf("First and third are the same")
	}
}

func TestLocator_Set_Errors(t *testing.T) {
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
			err := locator.Set(tt.factory)
			if err == nil {
				t.Error("No error.")
			}
			if !errors.Is(err, tt.expect) {
				t.Errorf("Wrong error. Got %v. Expect %v.", err, tt.expect)
			}
		})
	}
}

func TestLocator_Set_Sub(t *testing.T) {
	parent := New()
	parent.Set(newPairToken(1))

	child := parent.Sub()
	err := child.Set(newPairToken(2))
	if err != nil {
		t.Errorf("Set error %v.", err)
	}

	getBoth := func(parentValue interface{}, childValue interface{}) {
		err = parent.Get(parentValue)
		if err != nil {
			t.Errorf("Parent Get error %v.", err)
		}
		err = child.Get(childValue)
		if err != nil {
			t.Errorf("Child Get error %v.", err)
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

	parent.Set(func() int { return 27 })
	var parentInt, childInt int
	getBoth(&parentInt, &childInt)
	if parentInt != childInt {
		t.Errorf("Parent and child integer are not the same.")
	}
}

func TestLocator_Get(t *testing.T) {
	locator := New()
	locator.Set(newPairToken(1))

	var direct testToken
	err := locator.Get(&direct)
	if err != nil {
		t.Errorf("Get value returned %v.", err)
	}

	var byfun testToken
	err = locator.Get(func(tok testToken) { byfun = tok })
	if err != nil {
		t.Errorf("Get function returned %v.", err)
	}
	if !direct.Is(byfun) {
		t.Errorf("Direct and by-function are not the same.")
	}

	err = locator.Get(func(loc *Locator, tok testToken) {
		if loc != locator {
			t.Errorf("Wrong locator.")
		}
		byfun = tok
	})
	if !direct.Is(byfun) {
		t.Errorf("Direct and by-function are not the same.")
	}
}

func TestLocator_Get_Error(t *testing.T) {
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
			name:     "NotFound function",
			receptor: func(v int) error { return nil },
			expect:   NotFound,
		},
		{
			name:     "NotReceptor value",
			receptor: 42,
			expect:   NotReceptor,
		},
		{
			name:     "NotReceptor function",
			receptor: func(tok testToken) testToken { return tok },
			expect:   NotReceptor,
		},
		{
			name:     "Error from receptor",
			receptor: func(tok testToken) error { return customError },
			expect:   customError,
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
			locator.Set(newPairToken(1))

			if tt.factory != nil {
				err := locator.Set(tt.factory)
				if err != nil {
					t.Errorf("Error in set: %v.", err)
				}
			}

			err := locator.Get(tt.receptor)
			if !errors.Is(err, tt.expect) {
				t.Errorf("Wrong error. Got %v. Expect %v.", err, tt.expect)
			}
		})
	}
}

func TestLocator_GetFresh(t *testing.T) {
	locator := New()
	locator.Set(newPairToken(1))

	var first testToken
	err := locator.GetFresh(&first)
	if err != nil {
		t.Errorf("GetFresh returned %v.", err)
	}

	var singleton testToken
	err = locator.Get(&singleton)
	if err != nil {
		t.Errorf("Get returned %v.", err)
	}
	if singleton.Is(first) {
		t.Errorf("Singleton is the same as the first fresh.")
	}

	var second testToken
	err = locator.GetFresh(&second)
	if err != nil {
		t.Errorf("GetFresh returned %v.", err)
	}
	if singleton.Is(first) {
		t.Errorf("Singleton is the same as the second fresh.")
	}
	if singleton.Is(first) {
		t.Errorf("The second fresh is the same as the first fresh.")
	}

	var sub *Locator
	err = locator.GetFresh(&sub)
	if err != nil {
		t.Errorf("GetFresh returned %v.", err)
	}
	if sub == locator {
		t.Errorf("Sub is same as parent")
	}
}

func TestLocator_GetFresh_Error(t *testing.T) {
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
			locator.Set(newPairToken(1))

			if tt.factory != nil {
				err := locator.Set(tt.factory)
				if err != nil {
					t.Errorf("Error in set: %v.", err)
				}
			}

			err := locator.GetFresh(tt.receptor)
			if !errors.Is(err, tt.expect) {
				t.Errorf("Wrong error. Got %v. Expect %v.", err, tt.expect)
			}
		})
	}
}

func TestLocator_Inject(t *testing.T) {
	locator := New()
	locator.Set(newPairToken(1))

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
			expect:   NotFactory,
		},
		{
			name:    "No returned value",
			factory: func() {},
			expect:  NotFactory,
		},
		{
			name:     "Two returned values",
			factory:  func() (int, uint16) { return 1, 2 },
			receptor: &vInt,
			expect:   NotFactory,
		},
		{
			name:     "Three returned values",
			factory:  func() (int, error, uint32) { return 1, nil, 4 },
			receptor: &vInt,
			expect:   NotFactory,
		},
		{
			name:     "Not a pointer",
			factory:  func() int { return 42 },
			receptor: vInt,
			expect:   NotReceptor,
		},
		{
			name:     "Function as receptor",
			factory:  func() int { return 42 },
			receptor: func(v int) {},
			expect:   NotReceptor,
		},
		{
			name:     "Types mismatch",
			factory:  func() int { return 42 },
			receptor: &vToken,
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			locator := New()
			locator.Set(newPairToken(1))
			locator.Set(func() int { return 27 })

			err := locator.Inject(tt.factory, tt.receptor)
			if !errors.Is(err, tt.expect) {
				t.Errorf("Wrong error. Got %v. Expect %v.", err, tt.expect)
			}
		})
	}
}
