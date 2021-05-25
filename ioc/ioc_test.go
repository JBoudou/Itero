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

func TestInjector_New(t *testing.T) {
	injector := New()
	if injector == nil {
		t.Errorf("New returned nil")
	}

	var same *Injector
	err := injector.Get(&same)
	if err != nil {
		t.Errorf("Get returned error %v.", err)
	}
	if same != injector {
		t.Errorf("Get returned another injector")
	}
}

func TestInjector_Sub_New(t *testing.T) {
	injector := New()
	sub := injector.Sub()
	if sub == injector {
		t.Errorf("Sub returned the parent")
	}

	var same *Injector
	err := sub.Get(&same)
	if err != nil {
		t.Errorf("Get returned error %v.", err)
	}
	if same == injector {
		t.Errorf("Get returned the parent")
	}
	if same != sub {
		t.Errorf("Get returned another injector")
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

func TestInjector_Set(t *testing.T) {
	injector := New()

	err := injector.Set(newPairToken(1))
	if err != nil {
		t.Errorf("Set returned %v.", err)
	}

	var first testToken
	err = injector.Get(&first)
	if err != nil {
		t.Errorf("Get returned %v.", err)
	}
	if first.grp != 1 {
		t.Errorf("Wrong grp %d.", first.grp)
	}

	var second testToken
	err = injector.Get(&second)
	if err != nil {
		t.Errorf("Get returned %v.", err)
	}
	if !first.Is(second) {
		t.Errorf("First and second are not the same")
	}

	err = injector.Set(func() (testToken, error) { return newPairToken(2)(), nil })
	if err != nil {
		t.Errorf("Set returned %v.", err)
	}

	var third testToken
	err = injector.Get(&third)
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

func TestInjector_Set_Errors(t *testing.T) {
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

			injector := New()
			err := injector.Set(tt.factory)
			if err == nil {
				t.Error("No error.")
			}
			if !errors.Is(err, tt.expect) {
				t.Errorf("Wrong error. Got %v. Expect %v.", err, tt.expect)
			}
		})
	}
}

func TestInjector_Set_Sub(t *testing.T) {
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

func TestInjector_Get(t *testing.T) {
	injector := New()
	injector.Set(newPairToken(1))

	var direct testToken
	err := injector.Get(&direct)
	if err != nil {
		t.Errorf("Get value returned %v.", err)
	}

	var byfun testToken
	err = injector.Get(func(tok testToken) { byfun = tok })
	if err != nil {
		t.Errorf("Get function returned %v.", err)
	}
	if !direct.Is(byfun) {
		t.Errorf("Direct and by-function are not the same.")
	}

	err = injector.Get(func(inj *Injector, tok testToken) {
		if inj != injector {
			t.Errorf("Wrong injector.")
		}
		byfun = tok
	})
	if !direct.Is(byfun) {
		t.Errorf("Direct and by-function are not the same.")
	}
}

func TestInjector_Get_Error(t *testing.T) {
	var vInt int
	customError := errors.New("Custom")

	tests := []struct {
		name string
		factory interface{}
		receptor interface{}
		expect error
	}{
		{
			name: "NotFound value",
			receptor: &vInt,
			expect: NotFound,
		},
		{
			name: "NotFound function",
			receptor: func(v int) error { return nil },
			expect: NotFound,
		},
		{
			name: "NotReceptor value",
			receptor: 42,
			expect: NotReceptor,
		},
		{
			name: "NotReceptor function",
			receptor: func(tok testToken) testToken { return tok },
			expect: NotReceptor,
		},
		{
			name: "Error from receptor",
			receptor: func(tok testToken) error { return customError },
			expect: customError,
		},
		{
			name: "Error from factory",
			factory: func() (int, error) { return 0, customError },
			receptor: &vInt,
			expect: customError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			injector := New()
			injector.Set(newPairToken(1))
			
			if tt.factory != nil {
				err := injector.Set(tt.factory)
				if err != nil {
					t.Errorf("Error in set: %v.", err)
				}
			}

			err := injector.Get(tt.receptor)
			if !errors.Is(err, tt.expect) {
				t.Errorf("Wrong error. Got %v. Expect %v.", err, tt.expect)
			}
		})
	}
}

func TestInjector_GetFresh(t *testing.T) {
	injector := New()
	injector.Set(newPairToken(1))

	var first testToken
	err := injector.GetFresh(&first)
	if err != nil {
		t.Errorf("GetFresh returned %v.", err)
	}

	var singleton testToken
	err = injector.Get(&singleton)
	if err != nil {
		t.Errorf("Get returned %v.", err)
	}
	if singleton.Is(first) {
		t.Errorf("Singleton is the same as the first fresh.")
	}

	var second testToken
	err = injector.GetFresh(&second)
	if err != nil {
		t.Errorf("GetFresh returned %v.", err)
	}
	if singleton.Is(first) {
		t.Errorf("Singleton is the same as the second fresh.")
	}
	if singleton.Is(first) {
		t.Errorf("The second fresh is the same as the first fresh.")
	}

	var sub *Injector
	err = injector.GetFresh(&sub)
	if err != nil {
		t.Errorf("GetFresh returned %v.", err)
	}
	if sub == injector {
		t.Errorf("Sub is same as parent")
	}
}

func TestInjector_GetFresh_Error(t *testing.T) {
	var vInt int
	customError := errors.New("Custom")

	tests := []struct{
		name string
		factory interface{}
		receptor interface{}
		expect error
	}{
		{
			name: "NotFound value",
			receptor: &vInt,
			expect: NotFound,
		},
		{
			name: "NotReceptor value",
			receptor: 42,
			expect: NotReceptor,
		},
		{
			name: "NotReceptor function",
			receptor: func(tok testToken) {},
			expect: NotReceptor,
		},
		{
			name: "Error from factory",
			factory: func() (int, error) { return 0, customError },
			receptor: &vInt,
			expect: customError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			injector := New()
			injector.Set(newPairToken(1))
			
			if tt.factory != nil {
				err := injector.Set(tt.factory)
				if err != nil {
					t.Errorf("Error in set: %v.", err)
				}
			}

			err := injector.GetFresh(tt.receptor)
			if !errors.Is(err, tt.expect) {
				t.Errorf("Wrong error. Got %v. Expect %v.", err, tt.expect)
			}
		})
	}
}
