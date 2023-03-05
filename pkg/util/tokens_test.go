/*
 * Copyright (c) 2022-2023 Zander Schwid & Co. LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 */

package util_test

import (
	"github.com/codeallergy/sprintframework/pkg/util"
	"github.com/stretchr/testify/require"
	"math"
	"math/rand"
	"testing"
)

func TestLongId(t *testing.T) {

	id, err := util.GenerateLongId()
	require.NoError(t, err)

	value, err := util.DecodeLongId(id)
	require.NoError(t, err)

	require.Equal(t, id, util.EncodeLongId(value))

}

func TestShortId(t *testing.T) {

	for i := 0; i < 100; i++ {
		n := rand.Uint64() % uint64(math.Pow10(i/5))
		str := util.EncodeId(n)
		actual, err := util.DecodeId(str)
		require.NoError(t, err)
		require.Equal(t, n, actual)
	}

}

func TestShowId(t *testing.T) {
	num, _ := util.DecodeId("s00001")
	println(num)

	println(util.EncodeId(num+1))
}
