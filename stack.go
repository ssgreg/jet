package jet

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

// StackFunctions allows to specify the stop/exit order for the given functions
// and the ability to cancel them with the single given context.
// Notes:
// - Each function starts with it's own goroutine.
// - The given context cancels the last function (top of the stack).
// - Each function st
//
func StackFunctions(ctx context.Context, jobs ...func(context.Context) error) (err error) {
	topCtx, topCancel := context.WithCancel(ctx)

	return stackFunctions(topCtx, topCancel, jobs...)
}

func stackFunctions(ctx context.Context, cancel context.CancelFunc, jobs ...func(context.Context) error) (err error) {
	if len(jobs) == 0 {
		panic("stackFunctions called with empty jobs")
	}

	useInternalErrorFirst := false
	var internalErr error
	defer func() {
		// By default internal error will be skipped.
		errorToSkip := internalErr
		// If internal error occurred first or error is nil - use internal error.
		if (useInternalErrorFirst || err == nil) && internalErr != nil {
			errorToSkip = err
			err = internalErr
		}
		// The best we can do with both internal error and error is to
		// return first error and to log second error.
		if errorToSkip != nil {
			fmt.Fprintf(os.Stderr, "skipped: %s\n", errorToSkip.Error())
		}
	}()

	externalCtx := ctx

	if len(jobs) > 1 {
		var wg sync.WaitGroup
		defer wg.Wait()
		defer cancel()

		var useInternalErrorIf uint32
		defer func() {
			useInternalErrorFirst = atomic.CompareAndSwapUint32(&useInternalErrorIf, 1, 0)
		}()

		// Separate context per each function is needed to stop functions one-by-one.
		var externalCancel context.CancelFunc
		externalCtx, externalCancel = context.WithCancel(context.Background())
		defer externalCancel()

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer externalCancel()
			defer atomic.StoreUint32(&useInternalErrorIf, 1)
			internalErr = stackFunctions(ctx, cancel, jobs[1:]...)
		}()
	}

	return jobs[0](externalCtx)
}