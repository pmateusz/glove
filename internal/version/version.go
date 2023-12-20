/*
 * Copyright 2023 The Glove Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.
 */

package version

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type Version struct {
	Major  int
	Minor  int
	Patch  int
	Suffix string
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d%s", v.Major, v.Minor, v.Patch, v.Suffix)
}

var versionPattern = regexp.MustCompile(`(?P<major>\d+)\.(?P<minor>\d+)\.(?P<patch>\d+)(?P<suffix>\S+)?`)

func mustParseInt(value string) int {
	number, err := strconv.Atoi(value)
	if err != nil {
		panic(err)
	}
	return number
}

var ErrInvalidFormat = errors.New("version: invalid format")

func ParseVersion(value string) (Version, error) {
	groups := versionPattern.FindStringSubmatch(value)
	if len(groups) < 5 {
		return Version{}, ErrInvalidFormat
	}

	return Version{
		Major:  mustParseInt(groups[1]),
		Minor:  mustParseInt(groups[2]),
		Patch:  mustParseInt(groups[3]),
		Suffix: groups[4],
	}, nil
}
