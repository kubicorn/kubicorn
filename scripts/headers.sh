#!/usr/bin/env bash

read -r -d '' LICENSE <<EOF
// Copyright © 2017 The Kubicorn Authors\x0a//\x0a// Licensed under the Apache License, Version 2.0 (the "License");\x0a// you may not use this file except in compliance with the License.\x0a// You may obtain a copy of the License at\x0a//\x0a//\ \ \ \ \ http://www.apache.org/licenses/LICENSE-2.0\x0a//\x0a// Unless required by applicable law or agreed to in writing, software\x0a// distributed under the License is distributed on an "AS IS" BASIS,\x0a// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\x0a// See the License for the specific language governing permissions and\x0a// limitations under the License.
EOF

FILES=$(find . -name "*.go" -not -path "./vendor/*")

for FILE in $FILES; do
        CONTENT=$(head -n 13 $FILE)
        if [ "$CONTENT" != "$LICENSE" ]; then
		ESCAPED=$(echo $LICENSE | awk '{printf "%s\\n", $0}')
		sed -i "1s|^|${ESCAPED}\n|" $FILE
        fi
done
