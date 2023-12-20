/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package proxy

import (
	"fmt"
	"strings"
)

type Action int32

const (
	TunnelAction Action = iota
	BlockAction
	MITMAction
)

func ParseAction(actionName string) (Action, error) {
	if actionName == "" {
		return TunnelAction, nil
	}

	switch strings.ToUpper(actionName) {
	case "MITM":
		return MITMAction, nil
	case "BLOCK":
		return BlockAction, nil
	case "TUNNEL":
		return TunnelAction, nil
	default:
		return 0, fmt.Errorf("failed to parse action %q, supported actions are: block, tunnel or MITM", actionName)
	}
}
