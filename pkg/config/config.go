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

// Package config provides configuration values.
//
// Configuration values are read from a file containing a JSON encoded map from keys to values.
// This file is searched in the current working directory, and recursively in its parent directories.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/JBoudou/Itero/pkg/slog"
)

var (
	values  map[string]json.RawMessage
	valLock sync.RWMutex

	rec struct {
		logger   slog.Leveled
		filename string
		maxDepth int
	}
)

// Error returned when the key is not found in the configuration.
type KeyNotFound string

func (self KeyNotFound) Error() string {
	return fmt.Sprintf("Key %s not found", string(self))
}

// ReadFile reads a JSON file and stores the recorded values for subsequent calls to Value and ValueOr.
// This method must be called once before Value or ValueOr are called.
// When called multiple times, only the values from the last read file are available.
func ReadFile(logger slog.Leveled, filename string, maxDepth int) (baseDir string, err error) {
	rec.logger = logger
	rec.filename = filename
	rec.maxDepth = maxDepth
	return readFile()
}

func readFile() (baseDir string, err error) {
	rec.logger.Log("Loading configuration")

	baseDir, err = FindFileInParent(rec.filename, rec.maxDepth)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		rec.logger.Errorf("No configuration file ./%s found! You must create it.", rec.filename)
		return
	}
	in, err := os.Open(filepath.Join(baseDir, rec.filename))
	if err != nil {
		return
	}
	defer in.Close()

	err = Read(in)
	return
}

// Read is a low-level function that reads the JSON configuration map directly from the given
// Reader.
func Read(in io.Reader) (err error) {
	decoder := json.NewDecoder(in)
	var tmp map[string]json.RawMessage
	err = decoder.Decode(&tmp)
	if err != nil {
		return
	}

	valLock.Lock()
	values = tmp
	valLock.Unlock()
	return
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

// FindFileInParent search a file with the given filename.
// The search starts in the current directory then explores recursively the parent directories. The
// search fails after maxdepth changes of directory, i.e., when maxdepth is zero the file is search
// only in the current directory.
func FindFileInParent(filename string, maxdepth int) (path string, err error) {
	path, err = os.Getwd()
	if err != nil {
		return
	}

	for i := 0; i <= maxdepth; i++ {
		var stat os.FileInfo
		if stat, err = os.Stat(filepath.Join(path, filename)); err == nil && !stat.IsDir() {
			return
		}
		path = filepath.Dir(path)
	}
	return "", os.ErrNotExist
}
