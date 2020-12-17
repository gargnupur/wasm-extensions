// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"testing"
	"time"

	"istio.io/proxy/test/envoye2e"
	"istio.io/proxy/test/envoye2e/driver"
	"istio.io/proxy/testdata"
)

func TestLocalRateLimitFilter(t *testing.T) {
	t.Parallel()
	var TestCases = []struct {
		name          string
		canonicalName string
		responseCode  int
	}{
		{"MatchConfig", "ratings", 429},
		{"DefaultConfig", "reviews", 200},
	}

	for _, tt := range TestCases {
		t.Run(tt.name, func(t *testing.T) {
			params := driver.NewTestParams(t, map[string]string{
				"LocalRateLimitConfigurationData": driver.LoadTestJSON("testdata/local_rate_limit/local_ratelimit_configuration_data.json"),
				"CanonicalName":                   tt.canonicalName,
			}, envoye2e.ProxyE2ETests)
			params.Vars["ServerHTTPFilters"] = driver.LoadTestData("testdata/local_rate_limit/local_rate_limit.yaml.tmpl")

			if err := (&driver.Scenario{
				[]driver.Step{
					&driver.XDS{},
					&driver.Update{
						Node: "server", Version: "0", Listeners: []string{string(testdata.MustAsset("listener/server.yaml.tmpl"))},
					},
					&driver.Envoy{
						Bootstrap:       params.FillTestData(string(testdata.MustAsset("bootstrap/server.yaml.tmpl"))),
						DownloadVersion: "1.7.0",
					},
					&driver.Sleep{Duration: 1 * time.Second},
					&driver.Repeat{N: 3,
						Step: &driver.HTTPCall{
							Port: params.Ports.ClientPort,
							Body: "hello, world!",
						},
					},
					&driver.Repeat{N: 2,
						Step: &driver.HTTPCall{
							Port:         params.Ports.ClientPort,
							ResponseCode: tt.responseCode,
						},
					},
					&driver.Sleep{Duration: 1 * time.Second},
					&driver.Repeat{N: 2,
						Step: &driver.HTTPCall{
							Port: params.Ports.ClientPort,
							Body: "hello, world!",
						},
					},
					&driver.Repeat{N: 2,
						Step: &driver.HTTPCall{
							Port:         params.Ports.ClientPort,
							ResponseCode: tt.responseCode,
						},
					},
				},
			}).Run(params); err != nil {
				t.Fatal(err)
			}
		})
	}
}
