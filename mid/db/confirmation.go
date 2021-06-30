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

package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/JBoudou/Itero/mid/salted"
)

type ConfirmationType string

const (
	ConfirmationTypeVerify ConfirmationType = "verify"
	ConfirmationTypePasswd ConfirmationType = "passwd"
)

// CreateConfirmation creates a new confirmation.
// The created confirmation will be valid only for the given duration, starting at the moment the
// confirmation is created.
// Currently, no check is made to ensure that the creation of the confirmation will succeed, and
// nothing is done when it doesn't.
func CreateConfirmation(ctx context.Context,
	user uint32, type_ ConfirmationType, duration time.Duration) (ret salted.Segment, err error) {

	ret, err = salted.New(0)
	if err != nil {
		return
	}

	const qCreate = `
	  INSERT INTO Confirmations (User, Salt, Type, Expires)
	  VALUE (?, ?, ?, ADDTIME(CURRENT_TIMESTAMP, ?))`
	var result sql.Result
	result, err = DB.ExecContext(ctx, qCreate,
		user, ret.Salt, type_, DurationToTime(duration))
	if err != nil {
		return
	}

	ret.Id, err = IdFromResult(result)
	return
}
