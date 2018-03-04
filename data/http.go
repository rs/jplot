package data

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/elgs/gojq"
)

type httpSource struct {
	c    chan res
	done chan struct{}
}

type res struct {
	jq  *gojq.JQ
	err error
}

// FromHTTP fetches data points from url every interval and keep size points.
func FromHTTP(url string, interval time.Duration, size int) *Points {
	h := httpSource{
		c:    make(chan res),
		done: make(chan struct{}),
	}
	go h.run(url, interval)
	return &Points{
		Size:   size,
		Source: h,
	}
}

func (h httpSource) run(url string, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	h.fetch(url)
	for {
		select {
		case <-t.C:
			h.fetch(url)
		case <-h.done:
			close(h.c)
			return
		}
	}
}

func (h httpSource) fetch(url string) {
	resp, err := http.Get(url)
	if err != nil {
		h.c <- res{err: err}
		return
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		h.c <- res{err: err}
		return
	}
	jq, err := gojq.NewStringQuery(string(b))
	h.c <- res{jq: jq, err: err}
}

func (h httpSource) Get() (*gojq.JQ, error) {
	res := <-h.c
	return res.jq, res.err
}

func (h httpSource) Close() error {
	close(h.done)
	return nil
}
