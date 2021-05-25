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
	"reflect"
	"sync"
)

var (
	NotFactory  = errors.New("Wrong type for a factory")
	NotReceptor = errors.New("Wrong type for a receptor")
	NotFound    = errors.New("No binding")
)

type binding struct {
	factory  reflect.Value
	instance reflect.Value
	mutex    sync.Mutex // on instance
}

type Injector struct {
	bindings map[reflect.Type]*binding
	parent   *Injector
}

func New() *Injector {
	return (&Injector{bindings: map[reflect.Type]*binding{}}).addMyself()
}

var errorType = reflect.TypeOf(errors.New).Out(0)

func (self *Injector) Set(factory interface{}) error {
	factType := reflect.TypeOf(factory)

	// Check type of factory
	if factType.Kind() != reflect.Func {
		return NotFactory
	}
	nbOut := factType.NumOut()
	if nbOut != 1 && (nbOut != 2 || !factType.Out(1).Implements(errorType)) {
		return NotFactory
	}

	// Record
	self.bindings[factType.Out(0)] = &binding{
		factory:  reflect.ValueOf(factory),
		instance: reflect.Value{},
	}
	return nil
}

func (self *Injector) Get(receptor interface{}) error {
	// TODO handle functions as receptor
	rcpType := reflect.TypeOf(receptor)
	
	switch rcpType.Kind() {

	case reflect.Func:
		nbOut := rcpType.NumOut()
		if nbOut > 1 ||
			(nbOut == 1 && rcpType.Out(0) != errorType) {
			return NotReceptor
		}

		result, err := self.call(reflect.ValueOf(receptor))
		if err != nil {
			return err
		}
		if len(result) > 0 && !result[0].IsNil() {
			return result[0].Interface().(error)
		}
		return nil

	case reflect.Ptr:
		instance, err := self.getInstance(rcpType.Elem())
		if err != nil {
			return err
		}
		reflect.ValueOf(receptor).Elem().Set(instance)
		return nil
	}

	return NotReceptor
}

func (self *Injector) GetFresh(receptor interface{}) error {
	ptrType := reflect.TypeOf(receptor)
	if ptrType.Kind() != reflect.Ptr {
		return NotReceptor
	}

	bind, err := self.findBinding(ptrType.Elem())
	if err != nil {
		return err
	}

	var instance reflect.Value
	instance, err = self.instanciate(bind)
	if err != nil {
		return err
	}

	reflect.ValueOf(receptor).Elem().Set(instance)
	return nil
}

func (self *Injector) Sub() *Injector {
	return (&Injector{bindings: map[reflect.Type]*binding{}, parent: self}).addMyself()
}

// Implementation //

func (self *Injector) addMyself() *Injector {
	self.bindings[reflect.TypeOf(self)] = &binding{
		factory:  reflect.ValueOf(func() *Injector { return self.Sub() }),
		instance: reflect.ValueOf(self),
	}
	return self
}

func (self *Injector) getInstance(bindType reflect.Type) (reflect.Value, error) {
	bind, err := self.findBinding(bindType)
	if err != nil {
		return reflect.Value{}, err
	}

	bind.mutex.Lock()
	defer bind.mutex.Unlock()

	if !bind.instance.IsValid() {
		bind.instance, err = self.instanciate(bind)
	}
	return bind.instance, err
}

func (self *Injector) findBinding(bindType reflect.Type) (*binding, error) {
	if ret, found := self.bindings[bindType]; found {
		return ret, nil
	}
	if self.parent == nil {
		return nil, NotFound
	}
	return self.parent.findBinding(bindType)
}

func (self *Injector) instanciate(bind *binding) (reflect.Value, error) {
	result, err := self.call(bind.factory)
	if err != nil {
		return reflect.Value{}, err
	}
	if len(result) == 0 {
		return reflect.Value{}, NotFactory
	}
	if len(result) > 1 && !result[1].IsNil() {
		var ok bool
		err, ok = result[1].Interface().(error)
		if !ok {
			err = NotFactory
		}
	}
	return result[0], err
}

func (self *Injector) call(fct reflect.Value) ([]reflect.Value, error) {
	fctType := fct.Type()
	argLen := fctType.NumIn()
	arguments := make([]reflect.Value, argLen)
	var err error
	for i := 0; i < argLen; i++ {
		arguments[i], err = self.getInstance(fctType.In(i))
		if err != nil {
			return nil, err
		}
	}
	return fct.Call(arguments), nil
}
