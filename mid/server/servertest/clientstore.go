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

package servertest

import (
	"errors"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var (
	Unimplemented = errors.New("Unimplemented")
)

// ClientStore is a sessions.Store that store sessions in client requests.
type ClientStore struct {
	Codecs []securecookie.Codec
}

func NewClientStore(keys ...[]byte) *ClientStore {
	return &ClientStore{Codecs: securecookie.CodecsFromPairs(keys...)}
}

// Get is currently not implemented.
func (self *ClientStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return nil, Unimplemented
}

func (self *ClientStore) New(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.NewSession(self, name), nil
}

func (self *ClientStore) Save(r *http.Request, w http.ResponseWriter,
	session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, self.Codecs...)
	if err != nil {
		return err
	}
	r.AddCookie(sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}
