/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2019 Zachary Schneider
 */

package zps

import (
	"errors"
	"sort"
)

type Pool struct {
	index  map[string]Solvables
	rindex map[string]Solvables
	frozen map[string]bool

	Solvables Solvables

	repos Repos
}

func NewPool(image *Repo, frozen map[string]bool, repos ...*Repo) (*Pool, error) {
	pool := &Pool{index: make(map[string]Solvables), rindex: make(map[string]Solvables), frozen: frozen}

	if pool.frozen == nil {
		pool.frozen = make(map[string]bool)
	}

	if image == nil {
		errors.New("zps.Pool: Image must not be nil, can be empty repository")
	}

	// Force set this for now
	image.Priority = -1
	pool.repos = append(pool.repos, image)

	if len(repos) > 0 {
		for _, repo := range repos {
			pool.repos = append(pool.repos, repo)
		}

		// Sort by priority
		sort.Sort(pool.repos)
	}

	pool.populate()

	return pool, nil
}

func (p *Pool) Contains(pkg *Pkg) bool {
	if _, ok := p.index[pkg.Name()]; ok {
		for _, candidate := range p.index[pkg.Name()] {
			if candidate.Version().EXQ(pkg.Version()) {
				return true
			}
		}
	}

	return false
}

func (p *Pool) Location(index int) *Repo {
	return p.repos[index]
}

func (p *Pool) Installed(req *Requirement) Solvable {
	if _, ok := p.index[req.Name]; ok {
		for index, candidate := range p.index[req.Name] {
			if candidate.Satisfies(req) && candidate.Priority() <= -1 {
				return p.index[req.Name][index]
			}
		}
	}

	return nil
}

func (p *Pool) Frozen(id string) bool {
	return p.frozen[id]
}

func (p *Pool) Image() Solvables {
	var image Solvables

	for index, solvable := range p.Solvables {
		if solvable.Priority() <= -1 {
			image = append(image, p.Solvables[index])
		}
	}

	return image
}

func (p *Pool) RepoCount() int {
	return len(p.repos)
}

func (p *Pool) WhatDepends(name string) Solvables {

	if _, ok := p.rindex[name]; ok {
		return p.rindex[name]
	}

	return nil
}

func (p *Pool) WhatProvides(req *Requirement) Solvables {
	var provides Solvables

	if _, ok := p.index[req.Name]; ok {
		for _, candidate := range p.index[req.Name] {
			// Exact equality will never satisfy provides for frozen entries
			// this will insure exact install requests will fail
			if candidate.Satisfies(req) || p.frozen[candidate.Id()] {
				provides = append(provides, candidate)
			}
		}
	}

	return provides
}

func (p *Pool) populate() {
	for index, repo := range p.repos {
		if repo.Enabled == false {
			continue
		}

		for _, solvable := range repo.Solvables() {
			solvable.SetPriority(repo.Priority)
			solvable.SetLocation(index)

			if p.frozen[solvable.Id()] && solvable.Priority() == -1 {
				solvable.SetPriority(-2)
			}

			p.Solvables = append(p.Solvables, solvable)
			p.index[solvable.Name()] = append(p.index[solvable.Name()], solvable)

			// install reverse index
			if repo.Priority == -1 {
				for _, req := range solvable.Requirements() {
					if req.Method == "depends" {
						p.rindex[req.Name] = append(p.rindex[req.Name], solvable)
					}
				}
			}

			// provides support
			for _, req := range solvable.Requirements() {
				if req.Method == "provides" {
					p.index[req.Name] = append(p.index[req.Name], solvable)
				}
			}
		}
	}

	sort.Sort(p.Solvables)
}
