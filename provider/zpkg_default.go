/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

/*
 * Copyright 2018 Zachary Schneider
 */

package provider

import (
	"context"
	"fmt"

	"github.com/chuckpreslar/emission"
	"github.com/fezz-io/zps/action"
)

type ZpkgDefault struct {
	*emission.Emitter
	zpkg *action.Zpkg

	phaseMap map[string]string
}

func NewZpkgDefault(zpkg action.Action, phaseMap map[string]string, emitter *emission.Emitter) Provider {
	return &ZpkgDefault{emitter, zpkg.(*action.Zpkg), phaseMap}
}

func (z *ZpkgDefault) Realize(ctx context.Context) error {
	switch z.phaseMap[Phase(ctx)] {
	default:
		z.Emit("action.info", fmt.Sprintf("%s %s", z.zpkg.Type(), z.zpkg.Key()))
		return nil
	}
}
