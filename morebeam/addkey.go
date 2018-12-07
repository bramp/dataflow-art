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
	"reflect"

	"github.com/apache/beam/sdks/go/pkg/beam"
	"github.com/apache/beam/sdks/go/pkg/beam/core/funcx"
	"github.com/apache/beam/sdks/go/pkg/beam/core/util/reflectx"
)

var (
	// Key function should map a single arg from X -> Y.
	sig = &funcx.Signature{
		Args:   []reflect.Type{beam.XType},
		Return: []reflect.Type{beam.YType},
	}
)

func init() {
	beam.RegisterType(reflect.TypeOf((*addKeyFn)(nil)).Elem())
}

// AddKey takes a PCollection<V> and returns a PCollection<KV<K, V>> where the key is calculated from keyFn.
func AddKey(s beam.Scope, keyFn interface{}, col beam.PCollection) beam.PCollection {
	s = s.Scope("AddKey")

	fn := reflectx.MakeFunc1x1(keyFn)
	keyType := fn.Type().Out(0)

	ss := funcx.Replace(sig, beam.XType, col.Type().Type())
	ss = funcx.Replace(ss, beam.YType, keyType)
	funcx.MustSatisfy(keyFn, ss)

	t := beam.TypeDefinition{
		Var: beam.YType, T: keyType,
	}

	return beam.ParDo(s, &addKeyFn{
		KeyFn: beam.EncodedFunc{Fn: fn},
	}, col, t)
}

type addKeyFn struct {
	KeyFn beam.EncodedFunc `json:"keyFn"`

	fn reflectx.Func1x1
}

func (f *addKeyFn) Setup() {
	f.fn = reflectx.ToFunc1x1(f.KeyFn.Fn)
}

func (f *addKeyFn) ProcessElement(elm beam.X) (beam.Y, beam.X) {
	return f.fn.Call1x1(elm), elm
}
