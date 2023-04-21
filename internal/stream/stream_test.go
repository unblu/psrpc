package stream

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"github.com/livekit/psrpc"
	"github.com/livekit/psrpc/internal"
	"github.com/livekit/psrpc/internal/rand"
	"github.com/livekit/psrpc/pkg/info"
)

func TestStream(t *testing.T) {
	s := NewStream[*internal.Request, *internal.Response](
		context.Background(),
		&info.RequestInfo{},
		rand.NewStreamID(),
		psrpc.DefaultClientTimeout,
		&testStreamAdapter{},
		nil,
		make(chan *internal.Response),
		make(map[string]chan struct{}),
	)

	var err atomic.Value

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		err.Store(s.Send(&internal.Request{}))
		wg.Done()
	}()

	var handleErr error
	go func() {
		for s.(*stream[*internal.Request, *internal.Response]).pending.Load() == 0 {
			time.Sleep(time.Millisecond)
		}
		handleErr = s.HandleStream(&internal.Stream{
			Body: &internal.Stream_Close{
				Close: &internal.StreamClose{
					Error: psrpc.ErrStreamClosed.Error(),
					Code:  string(psrpc.ErrStreamClosed.Code()),
				},
			},
		})
		wg.Done()
	}()
	wg.Wait()

	require.NoError(t, handleErr)
	require.EqualValues(t, psrpc.ErrStreamClosed, err.Load())
}

type testStreamAdapter struct {
	sendCalls  atomic.Int32
	closeCalls atomic.Int32
}

func (a *testStreamAdapter) Send(ctx context.Context, msg *internal.Stream) error {
	a.sendCalls.Inc()
	return nil
}

func (a *testStreamAdapter) Close(streamID string) {
	a.closeCalls.Inc()
}