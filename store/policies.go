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

package store

import (
	"fmt"
)

// Policy codes, different for each `Policy` created.
const (
	PolicyCodeBlock int = iota + 1
	PolicyCodeReserve
	PolicyCodeStick
)

type basePolicy struct {
	Name string `json:"id"`
	// Reason explains why this policy exists.
	Reason string `json:"reason"`
	// Issuer tells where this policy comes from.
	Issuer string `json:"issuer"`
	// Code is the code of the policy, usefull when the policy
	// is delivered to another context.
	Code int `json:"code"`
	// Desc describes how the policy acts.
	Desc string `json:"description"`
}

func (p basePolicy) ID() string {
	return p.Name
}

// GenPolicy is a general purpose policy that allows
// to configure the behaviour of the Accept function
// setting its AcceptFunc field.
//
// Used mainly in tests.
type GenPolicy struct {
	basePolicy
	Name string `json:"name"`

	// AcceptFunc is used as implementation
	// of Accept.
	AcceptFunc func(id, target string) bool `json:"-"`
}

func (p *GenPolicy) ID() string {
	return p.Name
}

// Accept implements Policy.
func (p *GenPolicy) Accept(id, target string) bool {
	return p.AcceptFunc(id, target)
}

// BlockPolicy blocks `SourceID`.
type BlockPolicy struct {
	basePolicy
	// Source that should be always refuted.
	SourceID string `json:"-"`
}

func NewBlockPolicy(issuer, sourceID string) *BlockPolicy {
	return &BlockPolicy{
		basePolicy: basePolicy{
			Name:   "block_" + sourceID,
			Issuer: issuer,
			Code:   PolicyCodeBlock,
			Desc:   fmt.Sprintf("source %v will no longer be used", sourceID),
		},
		SourceID: sourceID,
	}
}

// Accept implements Policy.
func (p *BlockPolicy) Accept(id, target string) bool {
	return id != p.SourceID
}

// ReservedPolicy is a Policy implementation. It is used to reserve a source
// for a specific connection target. Note that this does not mean that the
// others sources may not receive a connection to target, it just means that
// `SourceID` will not accept any other connection exept the ones that go to
// `Target`.
type ReservedPolicy struct {
	basePolicy
	SourceID string `json:"reserved_source_id"`
	Target   string `json:"target"`
}

func NewReservedPolicy(issuer, sourceID, target string) *ReservedPolicy {
	return &ReservedPolicy{
		basePolicy: basePolicy{
			Name:   fmt.Sprintf("reserve_%s_for_%s", sourceID, target),
			Issuer: issuer,
			Code:   PolicyCodeReserve,
			Desc:   fmt.Sprintf("source %v will only be used for connections to %s", sourceID, target),
		},
		SourceID: sourceID,
		Target:   target,
	}
}

// Accept implements Policy.
func (p *ReservedPolicy) Accept(id, target string) bool {
	if id == p.SourceID {
		return target == p.Target
	}
	return true
}

// AvoidPolicy is a Policy implementation. It is used to avoid giving
// connection to `Target` to `SourceID`.
type AvoidPolicy struct {
	basePolicy
	SourceID string `json:"avoid_source_id"`
	Target   string `json:"target"`
}

func NewAvoidPolicy(issuer, sourceID, target string) *AvoidPolicy {
	return &AvoidPolicy{
		basePolicy: basePolicy{
			Name:   fmt.Sprintf("avoid_%s_for_%s", sourceID, target),
			Issuer: issuer,
			Code:   PolicyCodeReserve,
			Desc:   fmt.Sprintf("source %v will not be used for connections to %s", sourceID, target),
		},
		SourceID: sourceID,
		Target:   target,
	}
}

// Accept implements Policy.
func (p *AvoidPolicy) Accept(id, target string) bool {
	if target == p.Target {
		return id != p.SourceID
	}
	return true
}

// HistoryQueryFunc describes the function that is used to query the bind
// history of an entity. It is called passing the connection target in question,
// and it returns the source identifier that is associated to it and true,
// otherwise false if none is found.
type HistoryQueryFunc func(string) (string, bool)

// StickyPolicy is a Policy implementation. It is used to make connections to
// some target be always bound with the same source.
type StickyPolicy struct {
	basePolicy
	BindHistory HistoryQueryFunc `json:"-"`
}

func NewStickyPolicy(issuer string, f HistoryQueryFunc) *StickyPolicy {
	return &StickyPolicy{
		basePolicy: basePolicy{
			Name:   "stick",
			Issuer: issuer,
			Code:   PolicyCodeStick,
			Desc:   "once a source receives a connection to a target, the following connections to the same target will be assigned to the same source",
		},
		BindHistory: f,
	}
}

// Accept implements Policy.
func (p *StickyPolicy) Accept(id, target string) bool {
	if hid, ok := p.BindHistory(target); ok {
		return id == hid
	}

	return true
}