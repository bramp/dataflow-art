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

package morebeam

import (
	"math/rand"

	"github.com/apache/beam/sdks/go/pkg/beam"
)

// Reshuffle takes a PCollection<A> and shuffles the data to help increase parallelism.
func Reshuffle(s beam.Scope, col beam.PCollection) beam.PCollection {
	// The shuffle groups by a random key, and then emits the resulting values.

	s = s.Scope("Reshuffle")
	col = beam.ParDo(s, func(x beam.X) (int, beam.X) {
		return rand.Int(), x
	}, col)
	col = beam.GroupByKey(s, col)
	return beam.ParDo(s, func(key int, values func(*beam.X) bool, emit func(beam.X)) {
		var x beam.X
		for values(&x) {
			emit(x)
		}
	}, col)
}
