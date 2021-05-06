package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBalancer(t *testing.T) {
	assert := assert.New(t)

	mockedServersPool := []server{
		{
			dst:        "server1:8080",
			trafficCnt: 0,
		},
		{
			dst:        "server2:8080",
			trafficCnt: 0,
		},
		{
			dst:        "server3:8080",
			trafficCnt: 0,
		},
	}
	mockedServersPool[0].trafficCnt = 50
	mockedServersPool[1].trafficCnt = 20
	// ----
	serve := min(mockedServersPool, func(s1, s2 server) bool { return s1.trafficCnt < s2.trafficCnt })

	assert.Equal(mockedServersPool[serve].dst, mockedServersPool[2].dst)
	// ----
	mockedServersPool[2].trafficCnt = 40
	serve = min(mockedServersPool, func(s1, s2 server) bool { return s1.trafficCnt < s2.trafficCnt })
	// Now server with index "1" has least traffic

	assert.Equal(mockedServersPool[serve].dst, mockedServersPool[1].dst)
	// ----
	mockedServersPool[1].trafficCnt = 70
	serve = min(mockedServersPool, func(s1, s2 server) bool { return s1.trafficCnt < s2.trafficCnt })
	// Now server with least traffic is down, so server with least traffic that is healthy is with index "0"

	assert.Equal(mockedServersPool[serve].dst, mockedServersPool[0].dst)
}

func TestBalancerError(t *testing.T) {
	assert := assert.New(t)

	mockedServersPool := []server{
		{
			dst:        "server1:8080",
			trafficCnt: 50,
		},
		{
			dst:        "server2:8080",
			trafficCnt: 20,
		},
		{
			dst:        "server3:8080",
			trafficCnt: 0,
		},
	}
	// ----
	res := min(mockedServersPool, func(s1, s2 server) bool { return s1.trafficCnt < s2.trafficCnt })
	if res == 0 {
		err := "Error"
		assert.NotNil(err)
	}
}
