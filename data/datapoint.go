package data

import (
	"sync"
)

type Points struct {
	// Size is the number of data point to store per metric.
	Size int

	points map[string][]float64
	last   map[string]float64
	mu     sync.Mutex
}

func (p *Points) Push(name string, value float64, counter bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	d := p.getLocked(name)
	if counter {
		var diff float64
		if last := p.last[name]; last > 0 {
			diff = value - last
		}
		p.last[name] = value
		value = diff
	}
	d = append(append(make([]float64, 0, p.Size), d[1:]...), value)
	p.points[name] = d
}

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
