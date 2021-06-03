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

package services

import ()

//
// Users
//

type CreateUserEvent struct {
	User uint32
}

//
// Polls
//

type CreatePollEvent struct {
	Poll uint32
}

// StartPollEvent is send when a poll is started.
type StartPollEvent struct {
	Poll uint32
}

type VoteEvent struct {
	Poll uint32
}

// NextRoundEvent is the type of events send when a new round starts.
type NextRoundEvent struct {
	Poll  uint32
	Round uint8
}

// ClosePollEvent is type of events send when a poll is closed.
type ClosePollEvent struct {
	Poll uint32
}

type DeletePollEvent struct {
	Poll         uint32
	Title        string
	Participants map[uint32]bool
}