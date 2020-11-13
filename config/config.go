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

// Package config provides configured values for all parts of the application.
//
// Configuration values are read from the file "config.json" in the initial
// directory of the running program. This means in particular that for test
// programs, the file "config.json" is searched in the package directory.
package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const configFileName = "config.json"

var values map[string]json.RawMessage

// Error returned when the key is not found in the configuration.
type KeyNotFound string

func (self KeyNotFound) Error() string {
	return fmt.Sprintf("Key %s not found", string(self))
}

func init() {
	in, err := os.Open(configFileName)
	if err != nil {
		log.Fatalf("Unable to open config file %s: %v", configFileName, err)
	}

	decoder := json.NewDecoder(in)
	err = decoder.Decode(&values)
	if err != nil {
		log.Fatalf("Unable to parse JSON file %s: %v", configFileName, err)
	}
}

// Retrieve the value associated to the given key.
// Same rules than with json.Unmarshal applies to ret.
func Value(key string, ret interface{}) (err error) {
	raw, ok := values[key]
	if !ok {
		err = KeyNotFound(key)
		return
	}

	err = json.Unmarshal(raw, ret)
	return
}

// Same as Value except that if the key is not found then byDefault
// is stored as the new value for that key, and returned.
// Same rules than with json.Marshal applies to byDefault.
func ValueOr(key string, ret interface{}, byDefault interface{}) (err error) {
	raw, ok := values[key]
	if !ok {
		raw, err = json.Marshal(byDefault)
		if err != nil {
			return
		}
		values[key] = raw
	}

	err = json.Unmarshal(raw, ret)
	return
}

// Wrapper around Value to simplify retrieval of strings.
func String(key string) (ret string, err error) {
	err = Value(key, &ret)
	return
}

// Wrapper around ValueOr to simplify retrieval of strings.
func StringOr(key string, byDefault string) (ret string, err error) {
	err = ValueOr(key, &ret, &byDefault)
	return
}

// Wrapper around Value to simplify retrieval of ints.
func Int(key string) (ret int, err error) {
	err = Value(key, &ret)
	return
}

// Wrapper around ValueOr to simplify retrieval of ints.
func IntOr(key string, byDefault int) (ret int, err error) {
	err = ValueOr(key, &ret, &byDefault)
	return
}
