# Dataflow Art

Example of using [Apache Beam](https://beam.apache.org/documentation/sdks/go/) to extract color palette information from paintings.

Work in progress.

To Run:

```shell
go run main.go --runner=direct 
```

or

```shell
go run main.go \
  --art gs://${BUCKET?}/art/ \
  --index gs://${BUCKET?}/art/all_data_info.csv \
  --output gs://${BUCKET?}/output/ \
  --runner dataflow \
  --project ${PROJECT?} \
  --temp_location gs://${BUCKET?}/tmp/ \
  --staging_location gs://${BUCKET?}/staging/ \
  --worker_harness_container_image=apache-docker-beam-snapshots-docker.bintray.io/beam/go:20180515
```

## Licence (Apache 2)

*This is not an official Google product (experimental or otherwise), it is just code that happens to be owned by Google.*

```
Copyright 2018 Google LLC

Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```