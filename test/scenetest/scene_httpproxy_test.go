/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tests

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/alipay/sofa-mosn/cmd/mosn"
)

func GetServerAddr(s *httptest.Server) string {
	return strings.Split(s.URL, "http://")[1]
}

func ParseHTTPResponse(t *testing.T, req *http.Request) string {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("request failed %v\n", req)
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("read response body failed %v\n", req)
		return ""
	}
	re := regexp.MustCompile("\nServerName:[a-zA-Z0-9]+\n")
	bodys := strings.Split(
		strings.Trim(re.FindString(string(body)), "\n"), ":",
	)
	if len(bodys) < 2 {
		t.Errorf("response has no server name\n")
		return ""
	}
	return bodys[1]
}

func TestHttpProxy(t *testing.T) {
	//start httptest server
	cluster1Server := &HTTPServer{
		t:    t,
		name: "server1",
	}
	server1 := httptest.NewServer(cluster1Server)
	defer server1.Close()
	cluster2Server := &HTTPServer{
		t:    t,
		name: "server2",
	}
	server2 := httptest.NewServer(cluster2Server)
	defer server2.Close()
	//mesh config
	cluster1 := []string{GetServerAddr(server1)}
	cluster2 := []string{GetServerAddr(server2)}
	meshAddr := CurrentMeshAddr()
	meshConfig := CreateHTTPRouteConfig(meshAddr, [][]string{cluster1, cluster2})
	mesh := mosn.NewMosn(meshConfig)
	go mesh.Start()
	defer mesh.Close()
	time.Sleep(5 * time.Second) //wait mesh and server start
	makeRequest := func(header string, path string) *http.Request {
		req, err := http.NewRequest("GET", fmt.Sprintf("http://%s/%s", meshAddr, path), nil)
		if err != nil {
			t.Fatalf("create request error:%v\n", err)
		}
		req.Header.Add("service", header)
		return req
	}
	//cluster1
	if clustername := ParseHTTPResponse(t, makeRequest("cluster1", "")); clustername != cluster1Server.name {
		t.Errorf("expected %s, but got %s\n", cluster1Server.name, clustername)
	}
	//cluster2
	if clustername := ParseHTTPResponse(t, makeRequest("cluster2", "")); clustername != cluster2Server.name {
		t.Errorf("expected %s, but got %s\n", cluster2Server.name, clustername)
	}
	//cluster2 path
	if clustername := ParseHTTPResponse(t, makeRequest("cluster1", "test.htm")); clustername != cluster2Server.name {
		t.Errorf("expected %s, but got %s\n", cluster2Server.name, clustername)
	}
}
