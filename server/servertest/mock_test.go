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

package servertest

import (
	"context"
	"testing"
)

func TestMockResponse_SendJSON(t *testing.T) {
	t.Run("Without", func(t *testing.T){
		ttt := &testing.T{}
		MockResponse{T:ttt}.SendJSON(context.Background(), 0)
		if !ttt.Failed() {
			t.Errorf("SendJSON did not fail")
		}
	})
	t.Run("With", func(t *testing.T){
		called := 0
		mock := MockResponse{
			T: &testing.T{},
			JsonFct: func(t *testing.T, ctx context.Context, data interface{}) {
				if converted, ok := data.(int); ok && converted == 42 {
					called += 1
				} else {
					t.Errorf("Wrong data %v.", data)
				}
			},
		}
		mock.SendJSON(context.Background(), 42)
		if called != 1 {
			t.Errorf("Given function called %d times. Expect one.", called)
		}
	})
}
