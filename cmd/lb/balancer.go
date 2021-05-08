package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/EugeniaKol/design-practice-2/httptools"
	"github.com/EugeniaKol/design-practice-2/signal"
)

var (
	port       = flag.Int("port", 8090, "load balancer port")
	timeoutSec = flag.Int("timeout-sec", 3, "request timeout time in seconds")
	https      = flag.Bool("https", false, "whether backends support HTTPs")

	traceEnabled = flag.Bool("trace", false, "whether to include tracing information into responses")
)

type server struct {
	dst        string
	trafficCnt int64
}

var (
	timeout     = time.Duration(*timeoutSec) * time.Second
	serversPool = []server{
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
	up = make([]bool, len(serversPool))
)

func scheme() string {
	if *https {
		return "https"
	}
	return "http"
}

func health(dst string) bool {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	req, _ := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s://%s/health", scheme(), dst), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func forward(s server, rw http.ResponseWriter, r *http.Request) error {
	ctx, _ := context.WithTimeout(r.Context(), timeout)
	fwdRequest := r.Clone(ctx)
	fwdRequest.RequestURI = ""
	fwdRequest.URL.Host = s.dst
	fwdRequest.URL.Scheme = scheme()
	fwdRequest.Host = s.dst

	resp, err := http.DefaultClient.Do(fwdRequest)
	if err == nil {
		for k, values := range resp.Header {
			for _, value := range values {
				rw.Header().Add(k, value)
			}
		}
		if *traceEnabled {
			rw.Header().Set("lb-from", s.dst)
		}
		log.Println("№№№№№fwd", resp.StatusCode, resp.Request.URL)
		rw.WriteHeader(resp.StatusCode)
		defer resp.Body.Close()

		bytes, err := io.Copy(rw, resp.Body)
		s.trafficCnt += bytes
		log.Println("###########response body has ", bytes)

		if err != nil {
			log.Printf("Failed to write response: %s", err)
		}
		return nil
	} else {
		log.Printf("Failed to get response from %s: %s", s.dst, err)
		rw.WriteHeader(http.StatusServiceUnavailable)
		return err
	}
}

func min(ss []server, compare func(server, server) bool) (best int) {
	best = 0
	for i, s := range ss {
		if compare(s, ss[best]) && up[i] {
			best = i
		}
	}
	return
}

func main() {
	flag.Parse()

	for i, server := range serversPool {
		i := i
		server := server
		go func() {
			for range time.Tick(10 * time.Second) {
				state := health(server.dst)
				log.Println(server.dst, state)
				up[i] = state
			}
		}()
	}

	frontend := httptools.CreateServer(*port, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// TODO: Рееалізуйте свій алгоритм балансувальника.
		serverIndex := min(serversPool, func(s1, s2 server) bool { return s1.trafficCnt < s2.trafficCnt })
		forward(serversPool[serverIndex], rw, r)
	}))

	log.Println("Starting load balancer...")
	log.Printf("Tracing support enabled: %t", *traceEnabled)
	frontend.Start()
	signal.WaitForTerminationSignal()
}
