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
	"errors"
	"os"
	"path/filepath"
	"testing"
)

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

func TestFindFileInParent(t *testing.T) {
	tests := []struct {
		name   string
		pwd    string // from testdata
		file   string
		depth  int
		expect string // from testdata
		err    error
	}{
		{
			name:  "Not found",
			pwd:   "a/b/c/d",
			file:  "notfound.txt",
			depth: 3,
			err:   os.ErrNotExist,
		},
		{
			name:  "Direct",
			pwd:   "a/b/c/d",
			file:  "foo.txt",
			depth: 3,
			expect: "a/b/c/d/foo.txt",
		},
		{
			name:  "Last",
			pwd:   "a/b/c/d",
			file:  "foo.cfg",
			depth: 3,
			expect: "a/foo.cfg",
		},
		{
			name:  "Too deep",
			pwd:   "a/b/c/d",
			file:  "foo.ini",
			depth: 3,
			err:   os.ErrNotExist,
		},
	}

	for _, tt := range tests {
		// Not parallel because CWD is per thread and thread != goroutines.
		t.Run(tt.name, func(t *testing.T) {
			tt.pwd = filepath.FromSlash(tt.pwd)
			tt.expect = filepath.FromSlash(tt.expect)

			expectabs, _ := filepath.Abs(filepath.Join("testdata", tt.expect))
			origin, _ := os.Getwd()
			os.Chdir(filepath.Join("testdata", tt.pwd))
			defer os.Chdir(origin)

			got, err := FindFileInParent(tt.file, tt.depth)

			if tt.err == nil {
				if err != nil {
					t.Errorf("Got unexpected error %v.", err)
				} else {
					if got != tt.expect {
						gotabs, _ := filepath.Abs(got)
						if gotabs != expectabs {
							gotstat, _ := os.Stat(gotabs)
							expectstat, _ := os.Stat(expectabs)
							if !os.SameFile(gotstat, expectstat) {
								t.Errorf("Wrong result. Got %s. Expect %s.", got, tt.expect)
								t.Errorf("expectabs %s.", expectabs)
							}
						}
					}
				}
			} else {
				if !errors.Is(err, tt.err) {
					t.Errorf("Wrong error. Got %v. Expect %v.", err, tt.err)
				}
			}
		})
	}
}
