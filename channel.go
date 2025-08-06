package channelz

import (
	"context"
	"time"
)

type Channel[T any] struct {
	Ch chan T
}

func (c *Channel[T]) OrDone(ctx context.Context) chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		select {
		case <-ctx.Done():
			return
		case val, ok := <-c.Ch:
			if !ok {
				return
			}
			c.SendSafe(ctx, val)
		}
	}()
	return out
}

func (c *Channel[T]) SendSafe(ctx context.Context, val T) bool {
	select {
	case c.Ch <- val:
		return true
	case <-ctx.Done():
		return false
	}
}

func (c *Channel[T]) SendWithTimeout(val T, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return c.SendSafe(ctx, val)
}

func (c *Channel[T]) Duplicate() chan T {
	this := make(chan T, cap(c.Ch))
	ret := make(chan T, cap(c.Ch))
	go func() {
		dex := NewDemultiplexer[T]()
		oldCh := c.Ch
		c.Ch = this
		dex.Demultiplex(context.Background(), oldCh, this, ret)
		close(this)
	}()
	return ret
}
