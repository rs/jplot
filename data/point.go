package data

import (
	"fmt"
	"io"
	"sync"

	"github.com/elgs/gojq"
)

// Getter is the interface used by Points to get next points.
type Getter interface {
	io.Closer
	Get() (*gojq.JQ, error)
}

// Points is a series of Size data points gathered from Source.
type Points struct {
	// Size is the number of data point to store per metric.
	Size   int
	Source Getter

	points map[string][]float64
	last   map[string]float64
	mu     sync.Mutex
}

// Run get data from the source and capture metrics following specs.
func (p *Points) Run(specs []Spec) error {
	for {
		jq, err := p.Source.Get()
		if err != nil {
			return fmt.Errorf("input error: %v", err)
		}
		if jq == nil {
			break
		}
		for _, spec := range specs {
			for _, f := range spec.Fields {
				v, err := jq.Query(f.Name)
				if err != nil {
					return fmt.Errorf("cannot get %s: %v", f.Name, err)
				}
				n, ok := v.(float64)
				if !ok {
					return fmt.Errorf("invalid type %s: %T", f.Name, v)
				}
				p.push(f.ID, n, f.IsCounter)
			}
		}
	}
	return nil
}

func (p *Points) push(name string, value float64, counter bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	d := p.getLocked(name)
	if counter {
		var diff float64
		if last := p.last[name]; last > 0 && last < value {
			diff = value - last
		}
		p.last[name] = value
		value = diff
	}
	d = append(append(make([]float64, 0, p.Size), d[1:]...), value)
	p.points[name] = d
}

// Get gets the points vector for name.
func (p *Points) Get(name string) []float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.getLocked(name)
}

func (p *Points) getLocked(name string) []float64 {
	if p.points == nil {
		p.points = make(map[string][]float64, 1)
		p.last = make(map[string]float64)
	}
	d, found := p.points[name]
	if !found {
		d = make([]float64, p.Size)
		p.points[name] = d
	}
	return d
}

// Close calls Close on Source.
func (p *Points) Close() error {
	return p.Source.Close()
}
