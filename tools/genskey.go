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

package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
)

type GenSKey struct {}

func (self GenSKey) Cmd() string {
	return "genskey"
}

func (self GenSKey) String() string {
	return "Generate a pair of keys for the session, in JSON format."
}

func (self GenSKey) Run(args []string) {
	auth := make([]byte, 32)
	if _, err := rand.Read(auth); err != nil {
		panic(err)
	}
	enco := make([]byte, 16)
	if _, err := rand.Read(enco); err != nil {
		panic(err)
	}

	out, err := json.Marshal([][]byte{auth, enco})
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
}

func init() {
	AddCommand(GenSKey{})
}
