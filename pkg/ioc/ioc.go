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
// Its main purpose is to provide singleton services to other singleton services, during initial
// objects construction.
//
// The principle is that locators register factories to provide values to receptors. A factory is a
// constructor function for its output type. This function is called to construct a singleton
// service. It may have arguments that are dependencies, i.e., services to be provided by the
// locator. A receptor is either a pointer on a service to be located, or a function whose arguments
// are all services to be located.
//
// By default, service are considered to be singletons and factories are called at most once, when
// first needed. Method GetFresh allows to force a factory to be called.
//
// Service location is considered to be a singleton service, therefore any Locator provides itself
// to receptors for type *Locator. This allows receptor functions to call Get() for some services
// and GetFresh() for some others.
//
// Beware that cyclic dependencies are currently not detected by the package, resulting silently in
// infinite loops.
//
// Performance
//
// The package uses reflection, hence it is not lightning fast. You shoudl use it when the
// application starts to initialise long lasting objects.
//
// Parallelism
//
// Calls to Set() must all be done from the same goroutine, for instance in init() functions. All
// other methods can be called from any goroutine.
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
// the singleton value is the sublocator itself, and fresh value are sublocator of the sublocator.
// Calls to Set on the sublocator does not change the bindings on the parent.
func (self *Locator) Sub() *Locator {
	return (&Locator{bindings: map[reflect.Type]*binding{}, parent: self}).addMyself()
}

var errorType = reflect.TypeOf(errors.New).Out(0)

// Set registers a factory for a type.
// The factory must be a function, and the type of its first returned value is the type it is
// associated to. If the factory has a second returned value, it must be of type error. The factory
// cannot have a third returned value.
//
// The factory is called at most once to construct a singleton, when a call to Set() or Inject()
// depends on the associated type. The Locator provides singletons for all the arguments of the
// factory. This often results in other registered factories to be called. Cyclic dependencies are
// currently not detected and result in infinite loops.
func (self *Locator) Set(factory interface{}) error {
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

// Get provides singleton values to a receptor.
// The receptor must be either a pointer to store one singleton into, or a function, in which case
// it is called with all its arguments set to singletons provided by the Locator. If the receptor is
// a function, it must return nothing or an error.
//
// The call returns a NotFound error if there is no binding on the Locator for the requested type.
// It also returns an error if a called factory or the receptor returns an error.
func (self *Locator) Get(receptor interface{}) error {
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

// GetFresh constructs a new value using a registered factory.
// The receptor must be a pointer the new value is assigned to. The receptor of this method cannot
// be a function.
//
// Beware that the arguments given to the factory are still singleton values. If you want the
// factory to use fresh values, make it depend on *Locator and call GetFresh() directly from the
// factory.
func (self *Locator) GetFresh(receptor interface{}) error {
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

// Injects calls the given factory to produce a value for the given receptor.
// The constraints on the type of the factory are the same as for Set(). The receptor must be a
// pointer on a variable that can hold the first returned value of the factory. Functions are not
// allowed as receptor for this method.
// 
// The expression locator.Inject(factory, receptor) is roughly equivalent to
//
//     locator.Get(func(a A, b B, c C) { *receptor = factory(a,b,c) })
//
// differing only in the handling of errors and in the fact that the caller of Inject() does not
// have to bother about the types and number of the arguments of the factory.
func (self *Locator) Inject(factory interface{}, receptor interface{}) error {
	// Check types
	outType, err := self.checkFactory(factory)
	if err != nil {
		return err
	}
	rcpType := reflect.TypeOf(receptor)
	if rcpType.Kind() != reflect.Ptr || !outType.AssignableTo(rcpType.Elem()) {
		return NotReceptor
	}

	// Call
	instance, err := self.instanciate(reflect.ValueOf(factory))
	if err != nil {
		return err
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

func (self *Locator) getInstance(bindType reflect.Type) (reflect.Value, error) {
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
		return nil, NotFound
	}
	return self.parent.findBinding(bindType)
}

func (self *Locator) instanciate(factory reflect.Value) (reflect.Value, error) {
	result, err := self.call(factory)
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
func (self *Locator) call(fct reflect.Value) ([]reflect.Value, error) {
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
