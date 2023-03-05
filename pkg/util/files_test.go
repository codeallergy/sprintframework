/**
  Copyright (c) 2022 Zander Schwid & Co. LLC. All rights reserved.
*/

package util_test

import (
	"github.com/codeallergy/sprintframework/pkg/util"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestParseFileMode(t *testing.T) {

	knownModes := map[string]os.FileMode{
		"-rwxrwxr-x":    os.FileMode(0775),
		"-rw-rw-r--":    os.FileMode(0664),
		"-rw-rw-rw-":    os.FileMode(0666),
		"-rwxrwx---":    os.FileMode(0770),
	}

	for expected, mode := range knownModes {

		str := mode.String()
		require.Equal(t, expected, str)

		actual := util.ParseFileMode(str)
		require.Equal(t, mode, actual, mode.String())

	}

}
