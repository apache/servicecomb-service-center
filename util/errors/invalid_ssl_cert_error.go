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
package errors

type InvalidSSLCert struct {
	URL    string
	Reason string
}

func NewInvalidSSLCert(url, reason string) *InvalidSSLCert {
	return &InvalidSSLCert{
		URL:    url,
		Reason: reason,
	}
}

func (err *InvalidSSLCert) Error() string {
	message := "Received invalid SSL certificate from " + err.URL
	if err.Reason != "" {
		message += " - " + err.Reason
	}
	return message
}
