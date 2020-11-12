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
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const configFileName = "config.json"

var values map[string]json.RawMessage

type KeyNotFound string

func (self KeyNotFound) Error() string {
	return fmt.Sprintf("Key %s not found", string(self))
}

func init() {
	pwd, err := os.Getwd()
	log.Println(pwd)
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

func String(key string) (ret string, err error) {
	raw, ok := values[key]
	if !ok {
		err = KeyNotFound(key)
		return
	}

	err = json.Unmarshal(raw, &ret)
	return
}

func StringOr(key string, byDefault string) (ret string, err error) {
	raw, ok := values[key]
	if !ok {
		values[key], err = json.Marshal(&byDefault)
		ret = byDefault
		return
	}

	err = json.Unmarshal(raw, &ret)
	return
}
