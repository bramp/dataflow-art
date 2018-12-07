// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package morebeam provides additional functions useful when building
// Apache Beam pipelines.
package morebeam

import "strings"

// TODO
// * Create something similar to AddKey, for mapping just the values.

// Join is similar to path.Join but safe to use on URLs or filepaths.
func Join(elem ...string) string {
	if len(elem) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(elem[0])
	for i, e := range elem[1:] {
		if !strings.HasSuffix(elem[i], "/") {
			sb.WriteRune('/')
		}
		sb.WriteString(e)
	}
	return sb.String()
}
