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
package net

import (
	//"code.google.com/p/go.net/websocket"
	"crypto/x509"
	"github.com/servicecomb/service-center/util/errors"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Config struct {
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
}

func TimeoutDialer(config *Config) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, config.ConnectTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(config.ReadWriteTimeout))
		return conn, nil
	}
}

func newHttpClient() *http.Client {
	return &http.Client{
		CheckRedirect: PrepareRedirect,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 200,
			Dial: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
		},
	}
}

func PrepareRedirect(req *http.Request, via []*http.Request) error {
	if len(via) > 1 {
		return errors.New("stopped after 1 redirect")
	}

	prevReq := via[len(via)-1]
	req.Header.Set("Authorization", prevReq.Header.Get("Authorization"))
	return nil
}

func WrapNetworkErrors(host string, err error) error {
	var innerErr error
	switch typedErr := err.(type) {
	case *url.Error:
		innerErr = typedErr.Err
		//case *websocket.DialError:
		//	innerErr = typedErr.Err
	}

	if innerErr != nil {
		switch typedErr := innerErr.(type) {
		case x509.UnknownAuthorityError:
			return errors.NewInvalidSSLCert(host, "unknown authority")
		case x509.HostnameError:
			return errors.NewInvalidSSLCert(host, "not valid for the requested host")
		case x509.CertificateInvalidError:
			return errors.NewInvalidSSLCert(host, "")
		case *net.OpError:
			return typedErr.Err
		}
	}

	return errors.NewWithError("Error performing request", err)

}
