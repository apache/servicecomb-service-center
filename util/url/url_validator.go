//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package urlvalidator

import (
	"net/url"
	"regexp"
	"strings"
)

const (
	URL     string = `^(([1-9]\d?|1\d\d|2[01]\d|22[0-3])(\.(1?\d{1,2}|2[0-4]\d|25[0-5])){2}(\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-4])))(:(\d{1,5}))$`
	PATTERN string = `#%`
)

var (
	rxURL = regexp.MustCompile(URL)
)
var (
	rxURI = regexp.MustCompile(PATTERN)
)

// IsURL check if the string is an URL.
func IsURL(str string) bool {
	if str == "" || len(str) >= 2083 || len(str) <= 3 || strings.HasPrefix(str, ".") {
		return false
	}
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	if strings.HasPrefix(u.Host, ".") {
		return false
	}
	if u.Host == "" && (u.Path != "" && !strings.Contains(u.Path, ".")) {
		return false
	}
	return rxURL.MatchString(str)

}
func IsRequestURI(uri string) bool {
	if uri == "" || len(uri) >= 2048 || len(uri) <= 3 || strings.HasPrefix(uri, ".") {
		return false
	}
	if strings.HasSuffix(uri, ";") || strings.HasSuffix(uri, "&") || strings.HasSuffix(uri, "?") || strings.HasSuffix(uri, "+") || strings.HasSuffix(uri, "@") || strings.Contains(uri, "//") {
		return false
	}
	return !rxURI.MatchString(uri)
}
