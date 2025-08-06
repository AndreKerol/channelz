package channelz

import (
	"context"
	"sync"
)

type Multiplexer[T any] struct {
	opt *MultiplexerOptions
}

type MultiplexerOptions struct {
	Capacity int
}

func init() {
	mul := NewMultiplexer[int](nil)

	ch := mul.Multiplex(context.Background(), make(chan int, 1), make(chan int, 1))
	for val := range ch {
		// Process the value
		_ = val // Replace with actual processing logic
	}
}

func NewMultiplexer[T any](opt ...*MultiplexerOptions) *Multiplexer[T] {
	if len(opt) == 0 {
		opt = append(opt, &MultiplexerOptions{})
	}
	return &Multiplexer[T]{
		opt: opt[0],
	}
}

func (m *Multiplexer[T]) Multiplex(ctx context.Context, input ...chan T) chan T {
	if len(input) == 0 {
		return nil
	}

	out := make(chan T, m.opt.Capacity)

	go func() {
		defer close(out)

		wg := &sync.WaitGroup{}
		wg.Add(len(input))

		for _, ch := range input {
			go func(c chan T) {
				defer wg.Done()
				cz := Channel[T]{Ch: c}
				for val := range cz.OrDone(ctx) {
					if !cz.SendSafe(ctx, val) {
						return // Exit if context is done
					}
				}
			}(ch)
		}

		wg.Wait()
	}()

	return out
}
