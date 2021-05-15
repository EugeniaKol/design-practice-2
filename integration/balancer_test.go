package integration

import (
	"flag"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

//const baseAddress = "http://localhost:8090"
var target = flag.String("target", "http://localhost:8090", "request target")

var client = http.Client{
	Timeout: 3 * time.Second,
}

func TestBalancer(t *testing.T) {
	flag.Parse()
	var server string
	for i := 0; i < 10; i++ {
		//		url := fmt.Sprintf("%s/api/v1/some-data", baseAddress)
		t.Log(fmt.Sprintf("Sending request to %s", *target))
		resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", *target))

		if err != nil {
			t.Error(err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Error(fmt.Sprintf("Response code: %d", resp.StatusCode))
		}

		t.Logf("response from [%s]", resp.Header.Get("lb-from"))
		if i == 0 {
			server = resp.Header.Get("lb-from")
		} else {
			require.Equal(t, server, resp.Header.Get("lb-from"))
		}
	}
}

func BenchmarkBalancer(b *testing.B) {
	flag.Parse()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", *target))
		if err != nil {
			b.Error(err)
		}
		if resp.StatusCode != http.StatusOK {
			b.Error(fmt.Sprintf("Response code: %d", resp.StatusCode))
		}
	}
}
