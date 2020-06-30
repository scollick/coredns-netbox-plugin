// Copyright 2020 Oz Tiram <oz.tiram@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package netbox

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	clog "github.com/coredns/coredns/plugin/pkg/log"
)

type Record struct {
	Address  string `json:"address"`
	HostName string `json:"dns_name,omitempty"`
}

type RecordsList struct {
	Records []Record `json:"results"`
}

var stupidCache = make(map[string]string)

func query(url, token, dns_name string) string {

	clog.Debug(stupidCache)

	records := RecordsList{}
	client := &http.Client{}
	var resp *http.Response
	clog.Debug("Querying ", fmt.Sprintf("%s/?dns_name=%s", url, dns_name))
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/?dns_name=%s", url, dns_name), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", token))

	for i := 1; i <= 10; i++ {
		resp, err = client.Do(req)

		if err != nil {
			clog.Fatalf("HTTP Error %v", err)
		}

		if resp.StatusCode == http.StatusOK {
			break
		}

		time.Sleep(1 * time.Second)
	}
	// TODO: check that we got status code 200
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		clog.Fatalf("Error reading body %v", err)
	}

	jsonAns := string(body)
	err = json.Unmarshal([]byte(jsonAns), &records)
	if err != nil {
		clog.Fatalf("could not unmarshal response %v", err)
	}

	if len(records.Records) == 0 {
		clog.Info("Recored not found in", jsonAns)
		return ""
	}

	ip_address := strings.Split(records.Records[0].Address, "/")[0]

	stupidCache[dns_name] = ip_address

	return ip_address
}
