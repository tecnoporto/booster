// Copyright © 2019 KIM KeepInMind GmbH/srl
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"container/ring"
)

// Ring is a proxy struct around a container/ring.
// It forces to use Source as Value instead of bare interface{}.
type Ring struct {
	*ring.Ring
}

// NewRing creates a new ring with size n.
func NewRing(n int) *Ring {
	return &Ring{ring.New(n)}
}

func NewRingSources(ss ...Source) *Ring {
	r := NewRing(len(ss))
	for _, v := range ss {
		r.Set(v)
		r.Next()
	}

	return r
}

// Source retrives the Value of the ring at this position.
func (r *Ring) Source() Source {
	if v, ok := r.Ring.Value.(Source); ok {
		return v
	}
	return nil
}

// Set sets the Value of the ring at this position to s.
func (r *Ring) Set(s Source) {
	r.Ring.Value = s
}

// Do executes f on each value of the ring.
func (r *Ring) Do(f func(Source)) {
	r.Ring.Do(func(i interface{}) {
		if s, ok := i.(Source); ok {
			f(s)
			return
		}
		f(nil)
	})
}

func (r *Ring) Link(s *Ring) *Ring {
	r.Ring = r.Ring.Link(s.Ring)
	return r
}

func (r *Ring) Unlink(n int) *Ring {
	return &Ring{r.Ring.Unlink(n)}
}

func (r *Ring) Next() *Ring {
	r.Ring = r.Ring.Next()
	return r
}

func (r *Ring) Prev() *Ring {
	r.Ring = r.Ring.Prev()
	return r
}
