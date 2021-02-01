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

import { NavStepStatus, FinalNavTreeNode, LinearNavTreeNode, } from './navtree';


describe('NavTree', () => {
  it('may consist of a final node', () => {
    const final = new FinalNavTreeNode('segment', 'title');
    
    expect(final.parent).toBeUndefined();
    expect(final.next()).toBeUndefined();
    expect(final.isFinal).toBeTrue();

    const status = final.makeStatus();
    expect(status.current).toBe(0);
    expect(status.steps).toEqual(['title']);
    expect(status.mayHaveMore).toBeFalse();
  });

  it('may contain a linear node', () => {
    const leaf = new FinalNavTreeNode('final', 'Leaf');
    const root = new LinearNavTreeNode('base', 'Root', leaf);

    expect(leaf.parent).toBe(root);
    expect(leaf.next()).toBeUndefined();
    expect(leaf.isFinal).toBeTrue();
    expect(leaf.makeStatus()).toEqual(new NavStepStatus(1, ['Root', 'Leaf'], false));

    expect(root.parent).toBeUndefined();
    expect(root.next()).toBe(leaf);
    expect(root.isFinal).toBeFalse();
    expect(root.makeStatus()).toEqual(new NavStepStatus(0, ['Root', 'Leaf'], false));
  });
});
