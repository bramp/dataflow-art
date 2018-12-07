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

// Package csvio contains transforms for reading CSV files with https://github.com/gocarina/gocsv.
package csvio

// TODO
// * Create a CSV Writer

import (
	"context"
	"encoding/csv"
	"io"
	"reflect"
	"strings"

	"github.com/apache/beam/sdks/go/pkg/beam"
	"github.com/apache/beam/sdks/go/pkg/beam/io/filesystem"
	"github.com/apache/beam/sdks/go/pkg/beam/log"
	"github.com/gocarina/gocsv"
)

func init() {
	beam.RegisterFunction(expandFn)
	beam.RegisterType(reflect.TypeOf((*readFn)(nil)).Elem())
}

// Read reads a set of CSV files and returns the lines as a PCollection<T>.
// T is defined by the reflect.TypeOf( YourType{} ) with csv tags as descripted by
// https://github.com/gocarina/gocsv
func Read(s beam.Scope, glob string, t reflect.Type) beam.PCollection {
	s = s.Scope("csvio.Read")

	filesystem.ValidateScheme(glob)
	return read(s, t, beam.Create(s, glob))
}

func read(s beam.Scope, t reflect.Type, col beam.PCollection) beam.PCollection {
	files := beam.ParDo(s, expandFn, col)
	return beam.ParDo(s,
		&readFn{Type: beam.EncodedType{T: t}},
		files,
		beam.TypeDefinition{Var: beam.XType, T: t})
}

func expandFn(ctx context.Context, glob string, emit func(string)) error {
	if strings.TrimSpace(glob) == "" {
		return nil // ignore empty string elements here
	}

	fs, err := filesystem.New(ctx, glob)
	if err != nil {
		return err
	}
	defer fs.Close()

	files, err := fs.List(ctx, glob)
	if err != nil {
		return err
	}
	for _, filename := range files {
		emit(filename)
	}
	return nil
}

type readFn struct {
	Type beam.EncodedType
}

func (f *readFn) ProcessElement(ctx context.Context, filename string, emit func(beam.X)) error {
	log.Infof(ctx, "Reading from %v", filename)

	fs, err := filesystem.New(ctx, filename)
	if err != nil {
		return err
	}
	defer fs.Close()

	fd, err := fs.OpenRead(ctx, filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	r := csv.NewReader(fd)

	elem := reflect.New(f.Type.T).Elem().Interface()
	csv, err := gocsv.NewUnmarshaller(r, elem)
	if err != nil {
		return err
	}

	for {
		record, err := csv.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		emit(record)
	}
}
