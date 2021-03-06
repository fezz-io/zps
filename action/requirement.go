/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package action

import (
	"fmt"
	"strings"
)

type Requirement struct {
	Name      string `json:"name" hcl:"name,label"`
	Method    string `json:"method" hcl:"method"`
	Operation string `json:"operation,omitempty" hcl:"operation"`
	Version   string `json:"version,omitempty" hcl:"version,optional"`
}

func NewRequirement() *Requirement {
	return &Requirement{}
}

func (r *Requirement) Key() string {
	return r.Name
}

func (r *Requirement) Type() string {
	return "Requirement"
}

func (r *Requirement) Columns() string {
	return strings.Join([]string{
		strings.ToUpper(r.Type()),
		r.Name,
		r.Method,
		r.Operation,
		r.Version,
	}, "|")
}

func (r *Requirement) Id() string {
	return fmt.Sprint(r.Type(), ".", r.Key())
}

func (r *Requirement) Condition() *bool {
	return nil
}

func (r *Requirement) MayFail() bool {
	return false
}

func (r *Requirement) IsValid() bool {
	if r.Name != "" && r.Method != "" && r.Operation != "" && r.Version != "" {
		return true
	}

	return false
}
