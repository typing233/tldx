package checker

import (
	"context"
	"sync"
	"tldx/internal/rdap"
)

type Pool struct {
	client      *rdap.Client
	concurrency int
}

func NewPool(client *rdap.Client, concurrency int) *Pool {
	return &Pool{
		client:      client,
		concurrency: concurrency,
	}
}

func (p *Pool) Run(ctx context.Context, domains []string) <-chan Result {
	jobs := make(chan string, p.concurrency*2)
	results := make(chan Result, p.concurrency*2)

	// Dispatcher
	go func() {
		defer close(jobs)
		for _, d := range domains {
			select {
			case jobs <- d:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Workers
	var wg sync.WaitGroup
	for i := 0; i < p.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for domain := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}
				r := p.client.Check(ctx, domain)
				select {
				case results <- Result{
					Domain:    r.Domain,
					Available: r.Available,
					Error:     r.Error,
					Duration:  r.Duration,
				}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Closer
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}
