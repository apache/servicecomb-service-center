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
diff -u <(echo -n) <(find . -name "*.go" -not -path "./vendor/*" -not -path ".git/*" -not -path "./third_party/*" | grep -v _test | xargs gocyclo -over 16)
if [ $? == 0 ]; then
	echo "All function has less cyclomatic complexity..."
	exit 0
else
	echo "Functions/function has more cyclomatic complexity..."
	exit 1
fi
