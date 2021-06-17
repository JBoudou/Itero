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

// Package ioc provides simple inversion of control, based on service locator.
// Its main purpose is to provide singleton services during the initial objects construction.
//
// Overview
//
// Locators register factories to provide values to receptors. A factory is a constructor function
// for its output type. This function is called to construct a singleton service. It may have
// arguments that are dependencies, i.e., services to be provided by the locator. A receptor is
// either a pointer on a service to be located, or a function whose arguments are all services to be
// located.
//
// By default, service are considered to be singletons and factories are called at most once, when
// first needed. Methods Fresh and Refresh can be used to force a factory to be called.
//
// Service location is considered to be a singleton service, therefore any Locator provides itself
// to receptors of type *Locator. This allows receptor functions to call Inject() for some services
// and Fresh() for some others.
//
// Beware that cyclic dependencies are currently not detected by the package, resulting silently in
// infinite loops.
//
// Performance
//
// The package uses reflection, hence it is not lightning fast. You should use it when the
// application starts, to initialise long lasting objects.
//
// Parallelism
//
// Calls to Bind() and Refresh() must all be done from the same goroutine, for instance in init()
// functions. All other methods can be called from any goroutine.
package ioc

import (
	"errors"
	"reflect"
	"sync"
)

var (
	NotFactory  = errors.New("Wrong type for a factory")  // Factory's type is not handled.
	NotReceptor = errors.New("Wrong type for a receptor") // Receptor's type is not handled.
	NotFound    = errors.New("No binding")                // No factory for the requested type.
)

// Root is the base Locator.
// It may be used by packages to register default factories for the services they provide.
var Root *Locator = New()

type binding struct {
	factory  reflect.Value
	instance reflect.Value
	mutex    sync.Mutex // on instance
}

type Locator struct {
	bindings map[reflect.Type]*binding
	parent   *Locator
}

// New constructs a brand new Locator.
// The new Locator has bindings only of type *Locator, providing itself for singleton values, and
// calling Sub() for fresh values.
//
// Since packages may have registered factories on Root, you should probably use Root.Sub() instead.
func New() *Locator {
	return (&Locator{bindings: map[reflect.Type]*binding{}}).addMyself()
}

// Sub constructs a sub-locator of the current one.
// The sublocator initially has the same bindings as its parent, except for type *Locator for which
// the singleton value is the sublocator itself, and fresh values are sublocator of the sublocator.
// Calls to Bind on the sublocator does not change the bindings on the parent.
func (self *Locator) Sub() *Locator {
	return (&Locator{bindings: map[reflect.Type]*binding{}, parent: self}).addMyself()
}

var errorType = reflect.TypeOf(errors.New).Out(0)

// Bind registers a factory for a type.
// The factory must be a function, and the type of its first returned value is the type it is
// associated to. If the factory has a second returned value, it must be of type error. The factory
// cannot have a third returned value.
//
// The factory is called at most once to construct a singleton, when a call to Inject() (or Fresh())
// depends on the associated type. The Locator provides singletons for all the arguments of the
// factory. This often results in other registered factories to be called. Cyclic dependencies are
// currently not detected and result in infinite loops.
func (self *Locator) Bind(factory interface{}) error {
	outType, err := self.checkFactory(factory)
	if err != nil {
		return err
	}

	self.bindings[outType] = &binding{
		factory:  reflect.ValueOf(factory),
		instance: reflect.Value{},
	}
	return nil
}

// Inject retrieves singletons values from the locator.
//
// Its arguments are receptors that are considered in the order they are provided.
// Each receptor may be either a pointer or a function. Pointer receptors get their pointed value
// assigned to the deduced value. For function receptors, each of their arguments is set to a
// deduced value and the function is executed. If a function receptor returns values, these values
// are stored to be used as deduced value for subsequent receptors. If the last returned value has
// type error, this value is not stored and Inject fails with the returned error if it is not nil.
//
// The deduced value is the last stored value assignable to the requested type if there is one, or
// the singleton value registered for the requested type otherwise. For instance, the following code
// will store value 2 in the variable v.
//
//     var v int
//     Root.Inject(func() int {return 1}, func() int {return 2}, func() interface{} {return 3}, &v)
func (self *Locator) Inject(chain ...interface{}) error {
	// Check all types
	for _, receptor := range chain {
		if receptor == nil {
			return NotReceptor
		}
		rcpKind := reflect.TypeOf(receptor).Kind()
		if rcpKind != reflect.Func && rcpKind != reflect.Ptr {
			return NotReceptor
		}
	}

	store := make([]reflect.Value, 0, len(chain))
	for _, receptor := range chain {
		rcpType := reflect.TypeOf(receptor)
		switch rcpType.Kind() {

		case reflect.Func:
			result, err := self.call(reflect.ValueOf(receptor), store)
			if err != nil {
				return err
			}
			last := len(result)
			if last < 1 {
				continue
			}
			if err, ok := result[last-1].Interface().(error); ok {
				last -= 1
				if err != nil {
					return err
				}
			}
			store = append(store, result[:last]...)

		case reflect.Ptr:
			value, err := self.getInstance(rcpType.Elem(), store)
			if err != nil {
				return err
			}
			reflect.ValueOf(receptor).Elem().Set(value)

		}
	}

	return nil
}

// Fresh constructs a new value using a registered factory.
// The receptor must be a pointer the new value is assigned to. The receptor of this method cannot
// be a function.
//
// Beware that the arguments given to the factory are still singleton values. If you want the
// factory to use fresh values, make it depend on *Locator and call Fresh() directly from the
// factory.
func (self *Locator) Fresh(receptor interface{}) error {
	ptrType := reflect.TypeOf(receptor)
	if ptrType.Kind() != reflect.Ptr {
		return NotReceptor
	}

	bind, err := self.findBinding(ptrType.Elem())
	if err != nil {
		return err
	}

	var instance reflect.Value
	instance, err = self.instanciate(bind.factory)
	if err != nil {
		return err
	}

	reflect.ValueOf(receptor).Elem().Set(instance)
	return nil
}

// Refresh contructs a new value and use it as the new singleton value.
// Same constraints applies as with Fresh. Additionaly, this method cannot be called by multiple
// goroutines on the same Locator. However, Refresh can be called concurrently on different Locators
// with the same parent.
func (self *Locator) Refresh(receptor interface{}) error {
	ptrType := reflect.TypeOf(receptor)
	if ptrType.Kind() != reflect.Ptr {
		return NotReceptor
	}

	oldBind, err := self.findBinding(ptrType.Elem())
	if err != nil {
		return err
	}

	instance, err := self.instanciate(oldBind.factory)
	if err != nil {
		return err
	}

	self.bindings[ptrType.Elem()] = &binding{
		factory:  oldBind.factory,
		instance: instance,
	}

	reflect.ValueOf(receptor).Elem().Set(instance)
	return nil
}

// Implementation //

func (self *Locator) addMyself() *Locator {
	self.bindings[reflect.TypeOf(self)] = &binding{
		factory:  reflect.ValueOf(func() *Locator { return self.Sub() }),
		instance: reflect.ValueOf(self),
	}
	return self
}

// checkFactory check that its argument is a factory and returns the type it produces.
func (self *Locator) checkFactory(factory interface{}) (reflect.Type, error) {
	factType := reflect.TypeOf(factory)

	// Check type of factory
	if factType.Kind() != reflect.Func {
		return nil, NotFactory
	}
	nbOut := factType.NumOut()
	if nbOut != 1 && (nbOut != 2 || !factType.Out(1).Implements(errorType)) {
		return nil, NotFactory
	}

	return factType.Out(0), nil
}

func (self *Locator) getInstance(bindType reflect.Type, store []reflect.Value) (reflect.Value, error) {
	for i := len(store) - 1; i >= 0; i-- {
		if store[i].Type().AssignableTo(bindType) {
			return store[i], nil
		}
	}

	bind, err := self.findBinding(bindType)
	if err != nil {
		return reflect.Value{}, err
	}

	bind.mutex.Lock()
	defer bind.mutex.Unlock()

	if !bind.instance.IsValid() {
		bind.instance, err = self.instanciate(bind.factory)
	}
	return bind.instance, err
}

func (self *Locator) findBinding(bindType reflect.Type) (*binding, error) {
	if ret, found := self.bindings[bindType]; found {
		return ret, nil
	}
	if self.parent == nil {
		return nil, bindingError{bindType}
	}
	return self.parent.findBinding(bindType)
}

func (self *Locator) instanciate(factory reflect.Value) (reflect.Value, error) {
	result, err := self.call(factory, nil)
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

// call calls the given fonction with all its argument provided by the Locator.
// The method panic if its argument is not a function.
func (self *Locator) call(fct reflect.Value, store []reflect.Value) ([]reflect.Value, error) {
	fctType := fct.Type()
	argLen := fctType.NumIn()
	arguments := make([]reflect.Value, argLen)
	var err error
	for i := 0; i < argLen; i++ {
		arguments[i], err = self.getInstance(fctType.In(i), store)
		if err != nil {
			return nil, err
		}
	}
	return fct.Call(arguments), nil
}
