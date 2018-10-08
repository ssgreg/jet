package jet

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Shutdown(t *testing.T) {
	l, err := net.Listen("tcp", ":9796")
	require.NoError(t, err)

	s := http.Server{}

	go func() {
		time.Sleep(time.Second * 1)
		s.Shutdown(context.Background())
	}()

	err = RunServer(context.Background(), &s, l)
	require.NoError(t, err)
}

func Test_ShutdownWithContext(t *testing.T) {
	l, err := net.Listen("tcp", ":9797")
	require.NoError(t, err)

	serverCtx, serverCancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Second * 1)
		serverCancel()
	}()

	s := http.Server{}
	err = RunServer(serverCtx, &s, l)
	require.NoError(t, err)
}
