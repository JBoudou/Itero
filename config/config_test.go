// Itero - Online iterative vote application
// Copyright (C) 2020 Joseph Boudou
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

package config

import (
	"testing"
)

func TestString(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		const key = "string.foo"
		const expected = "foo"

		got, err := String(key)

		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if got != expected {
			t.Errorf("Got: %s. Expect: %s", got, expected)
		}
	})

	t.Run("not found", func(t *testing.T) {
		const key = "foobar"

		_, err := String(key)

		if err == nil {
			t.Fatalf("Key %s found", key)
		}
		knf, ok := err.(KeyNotFound)
		if !ok {
			t.Fatalf("Wrong type for error")
		}
		if string(knf) != key {
			t.Fatalf("Wrong key not found. Got: %s. Expect: %s.", string(knf), key)
		}
	})

	t.Run("or found", func(t *testing.T) {
		const key = "string.foo"
		const expected = "foo"

		got, err := StringOr(key, "other")

		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if got != expected {
			t.Errorf("Got: %s. Expect: %s", got, expected)
		}
	})

	t.Run("or not found", func(t *testing.T) {
		const key = "string.bar"
		const expected = "bar"

		_, err := String(key)
		if err == nil {
			t.Fatalf("Key %s found", key)
		}

		got, err := StringOr(key, expected)

		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if got != expected {
			t.Errorf("Got: %s. Expect: %s", got, expected)
		}
		got, err = String(key)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if got != expected {
			t.Errorf("Got: %s. Expect: %s", got, expected)
		}
	})
}

func TestInt(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		const key = "int.42"
		const expected = 42

		got, err := Int(key)

		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if got != expected {
			t.Errorf("Got: %d. Expect: %d", got, expected)
		}
	})

	t.Run("not found", func(t *testing.T) {
		const key = "foobar"

		_, err := Int(key)

		if err == nil {
			t.Fatalf("Key %s found", key)
		}
		knf, ok := err.(KeyNotFound)
		if !ok {
			t.Fatalf("Wrong type for error")
		}
		if string(knf) != key {
			t.Fatalf("Wrong key not found. Got: %s. Expect: %s.", string(knf), key)
		}
	})

	t.Run("or found", func(t *testing.T) {
		const key = "int.42"
		const expected = 42
		const other = 27

		got, err := IntOr(key, other)

		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if got != expected {
			t.Errorf("Got: %d. Expect: %d", got, expected)
		}
	})

	t.Run("or not found", func(t *testing.T) {
		const key = "int.27"
		const expected = 27

		_, err := Int(key)
		if err == nil {
			t.Fatalf("Key %s found", key)
		}

		got, err := IntOr(key, expected)

		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if got != expected {
			t.Errorf("Got: %d. Expect: %d", got, expected)
		}
		got, err = Int(key)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if got != expected {
			t.Errorf("Got: %d. Expect: %d", got, expected)
		}
	})
}

type myCompoundStruct struct {
	IntValue    int
	StringValue string
	FloatValue  float64
	Z           string `json:"other"`
	Array       [4]int
}

func TestValue(t *testing.T) {
	const key = "object"
	var expected = myCompoundStruct{42, "foo", 3.14, "bar", [4]int{2, 3, 5, 8}}

	const keyOther = "object.other"
	var other = myCompoundStruct{27, "flu", 2.14, "blu", [4]int{1, 3, 13, 75}}

	const keyPartial = "object.partial"
	var partial = myCompoundStruct{26, "foo", 3.14, "blu", [4]int{2, 3, 5, 8}}

	t.Run("found", func(t *testing.T) {
		var got myCompoundStruct
		err := Value(key, &got)

		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if got != expected {
			t.Errorf("Got: %v. Expect: %v", got, expected)
		}
	})

	t.Run("not found", func(t *testing.T) {
		var got myCompoundStruct
		err := Value(keyOther, &got)

		if err == nil {
			t.Fatalf("Key %s found", keyOther)
		}
		knf, ok := err.(KeyNotFound)
		if !ok {
			t.Fatalf("Wrong type for error")
		}
		if string(knf) != keyOther {
			t.Fatalf("Wrong key not found. Got: %s. Expect: %s.", string(knf), keyOther)
		}
	})

	t.Run("or found", func(t *testing.T) {
		var got myCompoundStruct
		err := ValueOr(key, &got, &other)

		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if got != expected {
			t.Errorf("Got: %v. Expect: %v", got, expected)
		}
	})

	t.Run("or not found", func(t *testing.T) {
		var got myCompoundStruct
		err := Value(keyOther, &got)
		if err == nil {
			t.Fatalf("Key %s found", keyOther)
		}

		err = ValueOr(keyOther, &got, &expected)

		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if got != expected {
			t.Errorf("Got: %v. Expect: %v", got, expected)
		}
		err = Value(keyOther, &got)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if got != expected {
			t.Errorf("Got: %v. Expect: %v", got, expected)
		}
	})

	t.Run("partial", func(t *testing.T) {
		got := partial
		err := Value(keyPartial, &got)

		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if got == partial {
			t.Errorf("Unchanged")
		}
		if got != expected {
			t.Errorf("Got: %v. Expect: %v", got, expected)
		}
	})
}
