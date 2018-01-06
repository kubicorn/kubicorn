// Copyright © 2017 The Kubicorn Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package defaults

import (
        "encoding/json"

        "github.com/kubicorn/kubicorn/apis/cluster"
)

func NewClusterDefaults(base *cluster.Cluster) *cluster.Cluster {
        // we need a deep clone, and also a copy of our pointer values
        // because the result of this function will be used for writing.
        // if we do not create copy's for the pointer values, we will
        // write into the original values!

        // this is ugly, because we ignore the errors, but otherwise
        // i'd have to change the signature of this func.
        // but i know that these structs are json serializable!
        var newcluster cluster.Cluster
        buf, _ := json.Marshal(base)
        json.Unmarshal(buf, &newcluster)
        return &newcluster
}

