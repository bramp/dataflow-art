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

// Example of using the csvio package.
package main

import (
	"context"
	"flag"
	"fmt"
	_ "image/jpeg"
	"reflect"

	"bramp.net/dataflow-art/morebeam/csvio"

	"github.com/apache/beam/sdks/go/pkg/beam"
	"github.com/apache/beam/sdks/go/pkg/beam/log"
	"github.com/apache/beam/sdks/go/pkg/beam/transforms/stats"
	"github.com/apache/beam/sdks/go/pkg/beam/x/beamx"
	"github.com/apache/beam/sdks/go/pkg/beam/x/debug"
)

type Painting struct {
	Artist  string `csv:"artist"`
	Title   string `csv:"title"`
	Year    int    `csv:"year"`
	NotUsed string `csv:"-"` // Ignored field
}

var (
	input = flag.String("input", "paintings.csv", "Input CSV file")
)

func extractFn(ctx context.Context, painting Painting) string {
	return painting.Artist
}

func formatFn(ctx context.Context, artist string, count int) string {
	return fmt.Sprintf("%s: %d", artist, count)
}

func main() {
	flag.Parse()
	beam.Init()

	ctx := context.Background()

	if *input == "" {
		log.Fatal(ctx, "No input provided")
	}

	p := beam.NewPipeline()
	s := p.Root()

	// Read the CSV file
	paintings := csvio.Read(s, *input, reflect.TypeOf(Painting{}))

	artists := beam.ParDo(s, extractFn, paintings)

	// Count the number of paintings by each artist.
	counts := stats.Count(s, artists)
	debug.Print(s, counts)

	if err := beamx.Run(ctx, p); err != nil {
		log.Fatalf(ctx, "Failed to execute job: %v", err)
	}
}
