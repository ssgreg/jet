package jethttp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
)

// RunServer starts the server synchronously with the given context, server
// and listener. The server can be gracefully shut down using the given context.
//
// RunServer returns nil in case of ErrServerClosed.
func RunServer(ctx context.Context, s *http.Server, l net.Listener) (err error) {
	var shutdownError error
	defer func() {
		if shutdownError != nil {
			if err == nil || err == http.ErrServerClosed {
				err = shutdownError
			} else {
				fmt.Fprintf(os.Stderr, "skipped: %s\n", err.Error())
			}
		}
		if err == http.ErrServerClosed {
			err = nil
		}
	}()

	var wg sync.WaitGroup
	defer wg.Wait()

	shutdownCtx, shutdownCancel := context.WithCancel(ctx)
	defer shutdownCancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-shutdownCtx.Done()
		shutdownError = s.Shutdown(context.Background())
	}()

	return s.Serve(l)
}
