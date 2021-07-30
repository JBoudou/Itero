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

// CreateUserEvent is sent when a new user has been created.
type CreateUserEvent struct {
	User uint32
}

// ReverifyEvent is sent when the user asks for its email address to be verified.
type ReverifyEvent struct {
	User uint32
}

// ForgotEvent is sent when a user requests to change its password.
type ForgotEvent struct {
	User uint32
}

//
// Polls
//

// CreatePollEvent is sents when a new poll has been created.
type CreatePollEvent struct {
	Poll uint32
}

// StartPollEvent is sent when a poll has started.
type StartPollEvent struct {
	Poll uint32
}

// VoteEvent is sent when a new ballot has been accepted for a poll.
type VoteEvent struct {
	Poll uint32
}

// NextRoundEvent is sent when a new round starts. It is not sent for the first round.
type NextRoundEvent struct {
	Poll  uint32
	Round uint8
}

// ClosePollEvent is sent when a poll has been marked as terminated.
type ClosePollEvent struct {
	Poll uint32
}

// DeletePollEvent is sent when a poll has been deleted.
type DeletePollEvent struct {
	Poll         uint32
	Title        string
	Participants map[uint32]bool
}
