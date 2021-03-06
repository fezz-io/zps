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
	"sort"
	"strings"
)

type OsArch struct {
	Os   string
	Arch string
}

type OsArches []*OsArch

func (slice OsArches) Len() int {
	return len(slice)
}

func (slice OsArches) Less(i, j int) bool {
	return slice[i].String() < slice[j].String()
}

func (slice OsArches) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func supportedPlatforms() map[string][]string {
	return map[string][]string{
		"any":     {"any", "arm64", "x68_64"},
		"darwin":  {"any", "x86_64"},
		"freebsd": {"any", "arm64", "x86_64"},
		"linux":   {"any", "arm64", "x86_64"},
	}
}

func Platforms() OsArches {
	var platforms OsArches
	for os, arches := range supportedPlatforms() {
		for _, arch := range arches {
			platforms = append(platforms, &OsArch{os, arch})
		}
	}

	sort.Sort(platforms)

	return platforms
}

func ExpandOsArch(osarch *OsArch) OsArches {
	var osArches OsArches

	osArches = append(osArches, osarch)
	osArches = append(osArches, &OsArch{"any", "any"})
	osArches = append(osArches, &OsArch{osarch.Os, "any"})
	osArches = append(osArches, &OsArch{"any", osarch.Arch})

	return osArches
}

func (oa *OsArch) String() string {
	return strings.Join([]string{oa.Os, oa.Arch}, "-")
}
