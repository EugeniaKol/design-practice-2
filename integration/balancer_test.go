package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func TestBalancer(t *testing.T) { TestingT(t) }

type IntegrationTest struct{}

var _ = Suite(&IntegrationTest{})

const baseAddress = "http://balancer:8090"

var client = http.Client{
	Timeout: 3 * time.Second,
}

func (s *IntegrationTest) TestBalancer(c *C) {
	// TODO: Реалізуйте інтеграційний тест для балансувальникка.
	serverPool := []string{
		"server1:8080",
		"server2:8080",
		"server3:8080",
	}
	authors := make(chan string, 10)
	for i := 0; i < 10; i++ {
		go func() {
			resp, err := client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
			if err != nil {
				c.Error(err)
			}
			respServer := resp.Header.Get("Lb-from")
			authors <- respServer
		}()
		time.Sleep(time.Duration(20) * time.Millisecond)
	}
	for i := 0; i < 10; i++ {
		auth := <-authors
		c.Assert(auth, Equals, serverPool[i%3])
	}
}

func (s *IntegrationTest) BenchmarkBalancer(c *C) {
	for i := 0; i < c.N; i++ {
		client.Get(fmt.Sprintf("%s/api/v1/some-data", baseAddress))
	}
	// TODO: Реалізуйте інтеграційний бенчмарк для балансувальникка.
}
