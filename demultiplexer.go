package channelz

import "context"

type Demultiplexer[T any] struct {
	opt *DemultiplexerOptions
}

type DemultiplexerOptions struct {
	Capacity int
}

func NewDemultiplexer[T any](opt ...*DemultiplexerOptions) *Demultiplexer[T] {
	if len(opt) == 0 {
		opt = append(opt, &DemultiplexerOptions{})
	}
	return &Demultiplexer[T]{
		opt: opt[0],
	}
}

func (d *Demultiplexer[T]) Demultiplex(ctx context.Context, input chan T, outputs ...chan T) {
	if len(outputs) == 0 {
		return
	}

	go func() {
		defer func() {
			for _, out := range outputs {
				close(out)
			}
		}()

		inputCh := Channel[T]{Ch: input}
		for val := range inputCh.OrDone(ctx) {
			for _, out := range outputs {
				outCh := Channel[T]{Ch: out}
				if !outCh.SendSafe(ctx, val) {
					return // Exit if context is done
				}
			}
		}
	}()
}
