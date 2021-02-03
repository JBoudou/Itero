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

/** Event about the current step in the navigation tree. */
export class NavStepStatus {
  constructor(
    /** Depth of the current node in the tree (root has depth zero). This is also the index in steps. */
    public current: number,
    /** Branch of the the tree containing the current node. */
    public steps: string[],
    /** Whether steps is a partial branch, up to the last undecided conditional node. */
    public mayHaveMore = false
  ) {};

  isFirst(): boolean {
    return this.current === 0;
  }

  isFinal(): boolean {
    return !this.mayHaveMore && this.current === this.steps.length - 1;
  }
}
