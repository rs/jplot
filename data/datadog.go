package data

import (
	"fmt"
	"time"

	"github.com/elgs/gojq"
	"gopkg.in/zorkian/go-datadog-api.v2"
)

type datadogSource struct {
	client *datadog.Client
	specs  []Spec
	c      chan res
	done   chan struct{}

	// state
	lastQueryTime int64
}

// FromDatadog fetches data from Datadog service (http://datadog.com).
func FromDatadog(apiKey, appKey string, specs []Spec, interval time.Duration, size int) *Points {
	client := datadog.NewClient(apiKey, appKey)
	s := &datadogSource{
		client:        client,
		specs:         specs,
		c:             make(chan res),
		done:          make(chan struct{}),
		lastQueryTime: time.Now().Unix(),
	}
	go s.run(interval)
	return &Points{
		Size:   size,
		Source: s,
	}
}

func (s *datadogSource) run(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			s.fetch()
		case <-s.done:
			close(s.c)
			return
		}
	}
}

func (s *datadogSource) fetch() {
	maxUpdateTimestamp := int64(s.lastQueryTime)
	dataPoints := make(map[string]datadog.DataPoint, len(s.specs))
	var err error
	for _, spec := range s.specs {
		for _, field := range spec.Fields {
			query := s.formatQuery(field)
			series, err := s.client.QueryMetrics(s.lastQueryTime, time.Now().Unix(), query)
			if err != nil {
				s.c <- res{err: err}
				return
			}
			if len(series) == 0 {
				s.c <- res{err: fmt.Errorf("no data for %s", field.Name)}
				return
			}
			endTs := int64(series[0].GetEnd() / 1000)
			if endTs > maxUpdateTimestamp {
				maxUpdateTimestamp = endTs
			}
			// assume the last data point is the latest
			dataPoints[field.ID] = series[0].Points[len(series[0].Points)-1]
		}
	}

	jq := gojq.NewQuery(dataPoints)
	s.c <- res{jq: jq, err: err}
}

func (s *datadogSource) formatQuery(field Field) string {
	querySuffix := ""
	if field.IsCounter {
		querySuffix = ".as_count()"
	}
	return fmt.Sprintf("%s%s", field.Name, querySuffix)
}

func (s *datadogSource) Get() (*gojq.JQ, error) {
	res := <-s.c
	return res.jq, res.err
}

func (s *datadogSource) Close() error {
	close(s.done)
	return nil
}
