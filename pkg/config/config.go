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
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const (
	configFileName = "config.json"
	maxDepth       = 2
)

var (
	values  map[string]json.RawMessage
	valLock sync.RWMutex
)

// Ok is true iff the configuration file has successfully been read.
// It may be false after init() has been called if no configuration file has been found.
var Ok bool

// Error returned when the key is not found in the configuration.
type KeyNotFound string

func (self KeyNotFound) Error() string {
	return fmt.Sprintf("Key %s not found", string(self))
}

func init() {
	Ok = readConfigFile()
}

func readConfigFile() bool {
	log.Println("Loading configuration")

	foundfile, err := FindFileInParent(configFileName, maxDepth)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		log.Printf("WARNING: no configuration file ./%s found! You must create it.", configFileName)
		log.Printf("To enable tests, there must be a configuration file (or link) in each package folder.")
		return false
	}
	in, err := os.Open(foundfile)
	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(in)
	var tmp map[string]json.RawMessage
	err = decoder.Decode(&tmp)
	if err != nil {
		panic(err)
	}

	valLock.Lock()
	values = tmp
	valLock.Unlock()
	return true
}

// Value retrieves the value associated with the given key.
// Same rules than with json.Unmarshal applies to ret.
func Value(key string, ret interface{}) (err error) {
	valLock.RLock()
	raw, ok := values[key]
	valLock.RUnlock()

	if !ok {
		err = KeyNotFound(key)
		return
	}

	err = json.Unmarshal(raw, ret)
	return
}

// ValueOr is similar to Value except that if the key is not found then byDefault
// is stored as the new value for that key, and returned.
// Same rules than with json.Marshal applies to byDefault.
func ValueOr(key string, ret interface{}, byDefault interface{}) (err error) {
	valLock.Lock()
	defer valLock.Unlock()
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

func FindFileInParent(filename string, maxdepth int) (path string, err error) {
	var cur string
	cur, err = os.Getwd()
	if err != nil {
		return
	}

	for i := 0; i <= maxdepth; i++ {
		path = filepath.Join(cur, filename)
		var stat os.FileInfo
		if stat, err = os.Stat(path); err == nil && !stat.IsDir() {
			return
		}
		cur = filepath.Dir(cur)
	}
	return "", os.ErrNotExist
}
