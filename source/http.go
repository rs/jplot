package source

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/elgs/gojq"
)

type HTTP struct {
	c    chan res
	done chan struct{}
}

type res struct {
	jq  *gojq.JQ
	err error
}

func NewHTTP(url string, interval time.Duration) HTTP {
	h := HTTP{
		c:    make(chan res),
		done: make(chan struct{}),
	}
	go h.run(url, interval)
	return h
}

func (h HTTP) run(url string, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	h.fetch(url)
	for {
		select {
		case <-t.C:
			h.fetch(url)
		case <-h.done:
			return
		}
	}
}

func (h HTTP) fetch(url string) {
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

func (h HTTP) Get() (*gojq.JQ, error) {
	res := <-h.c
	return res.jq, res.err
}

func (h HTTP) Close() error {
	close(h.done)
	return nil
}
