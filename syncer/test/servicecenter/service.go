package servicecenter

import (
	"net/http"
)

func (m *mockServer) ServiceExistence(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte(`{
    "serviceId": "4042a6a3e5a2893698ae363ea99a69eb63fc51cd"
}`))
}

func (m *mockServer) CreateService(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte(`{
    "serviceId": "4042a6a3e5a2893698ae363ea99a69eb63fc51cd"
}`))
}

func (m *mockServer) DeleteService(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte(`{
    "serviceId": "4042a6a3e5a2893698ae363ea99a69eb63fc51cc"
}`))
}
