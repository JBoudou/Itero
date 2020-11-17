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
  "fmt"
  "os"
)

var (
  commands map[string]Command = make(map[string]Command, 10)
)

type Command interface {
  Cmd() string
  String() string
  Run(args []string)
}

func AddCommand(cmd Command) {
  commands[cmd.Cmd()] = cmd
}

func usage() {
  fmt.Printf("Usage: %s <command> <options>, where <command> is one of:\n", os.Args[0])
  for _, cmd := range commands {
    fmt.Printf("\t%s\t%s\n", cmd.Cmd(), cmd.String())
  }
}

func main() {
  if len(os.Args) < 2 {
    usage()
    return
  }

  cmd, found := commands[os.Args[1]]
  if !found {
    fmt.Printf("Unknown command %s.\n", os.Args[1])
    return
  }

  cmd.Run(os.Args[2:])
}
