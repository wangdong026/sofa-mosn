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
	"sync"
	"testing"
	"time"

	"github.com/alipay/sofa-mosn/cmd/mosn"
	"github.com/alipay/sofa-mosn/pkg/protocol"
	"github.com/alipay/sofa-mosn/pkg/types"
)

func TestSofaRpc(t *testing.T) {
	sofaAddr := "127.0.0.1:8080"
	meshAddr := CurrentMeshAddr()
	server := NewUpstreamServer(t, sofaAddr, ServeBoltV1)
	server.GoServe()
	defer server.Close()
	meshConfig := CreateSimpleMeshConfig(meshAddr, []string{sofaAddr}, protocol.SofaRPC, protocol.SofaRPC)
	mesh := mosn.NewMosn(meshConfig)
	go mesh.Start()
	defer mesh.Close()
	time.Sleep(5 * time.Second) //wait mesh and server start
	//client
	client := &BoltV1Client{
		t:        t,
		ClientID: "testClient",
		Waits:    sync.Map{},
	}
	client.Connect(meshAddr)
	defer client.conn.Close(types.NoFlush, types.LocalClose)
	for i := 0; i < 20; i++ {
		client.SendRequest()
	}
	if !WaitMapEmpty(client.Waits, 20*time.Second) {
		t.Errorf("exists request no response\n")
	}
}
