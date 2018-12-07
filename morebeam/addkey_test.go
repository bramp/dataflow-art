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
	"context"

	"github.com/apache/beam/sdks/go/pkg/beam"
	"github.com/apache/beam/sdks/go/pkg/beam/x/beamx"
	"github.com/apache/beam/sdks/go/pkg/beam/x/debug"
)

func AddKeyExample() {
	ctx := context.Background()
	p, s := beam.NewPipelineWithRoot()

	values := beam.CreateList(s, []string{"andrew", "bob", "china"})

	// PCollection<string> -> // PCollection<KV<int, string>>
	keyvalues := AddKey(s, func(value string) int {
		return len(value)
	}, values)

	debug.Print(s, keyvalues)
	beamx.Run(ctx, p)

	// Output:
	// 6, "andrew"
	// 3, "bob"
	// 5, "china"
}
