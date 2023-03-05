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

package server_test

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

var robotsTxt = "\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x0a\x2d\x4e\x2d\xd2\x4d\x4c\x4f\xcd\x2b\xb1\x52\xd0\xe2\xe5\x72\xc9\x2c\x4e\xcc\xc9\xc9\x2f\xb7\x52\xd0\x4f\x2c\xc8\xd4\xe7\xe5\x02\x04\x00\x00\xff\xff\x25\xc9\xc7\x6c\x20\x00\x00\x00"

func TestGzipUnpack(t *testing.T) {

	fmt.Printf("compressed len %d\n", len(robotsTxt))

	var plain bytes.Buffer
	zr, err := gzip.NewReader(bytes.NewReader([]byte(robotsTxt)))
	require.NoError(t, err)

	n, err := io.Copy(&plain, zr)
	require.NoError(t, err)
	err = zr.Close()
	require.NoError(t, err)

	println(n)
	plaintText := string(plain.Bytes())
	println(plaintText)

}

