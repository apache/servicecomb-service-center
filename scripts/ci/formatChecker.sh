# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
# 
#   http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
diff -u <(echo -n) <(find . -name "*.go" -not -path "./vendor/*" -not -path ".git/*" | xargs gofmt -s -d)
if [ $? == 0 ]; then
	echo "Hurray....all code is formatted properly..."
	exit 0
else
	echo "There is issues's with the code formatting....please run go fmt on your code"
	exit 1
fi
