package jetfn

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_SingleSuccess(t *testing.T) {
	fired := false
	err := Stack(context.Background(), func(ctx context.Context) error {
		fired = true
		return nil
	})
	require.NoError(t, err)
	require.True(t, fired, "stacked function should be called")
}

func Test_SingleError(t *testing.T) {
	err := Stack(context.Background(), func(ctx context.Context) error {
		return errors.New("single_error")
	})
	require.EqualError(t, err, "single_error")
}

func Test_SingleCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Millisecond * 100)
		cancel()
	}()
	now := time.Now()
	err := Stack(ctx, func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	})
	require.NoError(t, err)
	require.WithinDuration(t, now.Add(time.Millisecond*100), time.Now(), time.Millisecond*20)
}

func Test_DoubleBothSuccess(t *testing.T) {
	fired1 := false
	fired2 := false
	err := Stack(context.Background(),
		func(ctx context.Context) error {
			fired1 = true
			return nil
		},
		func(ctx context.Context) error {
			fired2 = true
			return nil
		},
	)
	require.NoError(t, err)
	require.True(t, fired1, "stacked function1 should be called")
	require.True(t, fired2, "stacked function2 should be called")
}

func Test_DoubleFirstError(t *testing.T) {
	fired2 := false
	err := Stack(context.Background(),
		func(ctx context.Context) error {
			return errors.New("double_first_error")
		},
		func(ctx context.Context) error {
			fired2 = true
			return nil
		},
	)
	require.EqualError(t, err, "double_first_error")
	require.True(t, fired2, "stacked function2 should be called")
}

func Test_DoubleSecondError(t *testing.T) {
	fired1 := false
	err := Stack(context.Background(),
		func(ctx context.Context) error {
			fired1 = true
			return nil
		},
		func(ctx context.Context) error {
			return errors.New("double_second_error")
		},
	)
	require.EqualError(t, err, "double_second_error")
	require.True(t, fired1, "stacked function1 should be called")
}

func Test_DoubleBothErrors_GotSecond(t *testing.T) {
	err := Stack(context.Background(),
		func(ctx context.Context) error {
			time.Sleep(time.Millisecond * 100)
			return errors.New("double_first_error")
		},
		func(ctx context.Context) error {
			return errors.New("double_second_error")
		},
	)
	require.EqualError(t, err, "double_second_error")
}

func Test_DoubleBothErrors_GotFirst(t *testing.T) {
	err := Stack(context.Background(),
		func(ctx context.Context) error {
			return errors.New("double_first_error")
		},
		func(ctx context.Context) error {
			time.Sleep(time.Millisecond * 100)
			return errors.New("double_second_error")
		},
	)
	require.EqualError(t, err, "double_first_error")
}

func Test_DoubleCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Millisecond * 100)
		cancel()
	}()
	now := time.Now()
	var exitTime1 time.Time
	var exitTime2 time.Time
	err := Stack(ctx,
		func(ctx context.Context) error {
			<-ctx.Done()
			exitTime1 = time.Now()
			return nil
		},
		func(ctx context.Context) error {
			<-ctx.Done()
			exitTime2 = time.Now()
			return nil
		},
	)
	require.NoError(t, err)
	require.True(t, exitTime2.Before(exitTime1), "exitTime2 should go first because it on top of the stack")
	require.WithinDuration(t, now.Add(time.Millisecond*100), time.Now(), time.Millisecond*20)
}

func Test_DoubleCancel_Internal(t *testing.T) {
	now := time.Now()
	var exitTime1 time.Time
	var exitTime2 time.Time
	err := Stack(context.Background(),
		func(ctx context.Context) error {
			time.Sleep(time.Millisecond * 100)
			exitTime1 = time.Now()
			return nil
		},
		func(ctx context.Context) error {
			<-ctx.Done()
			exitTime2 = time.Now()
			return nil
		},
	)
	require.NoError(t, err)
	require.True(t, exitTime1.Before(exitTime2))
	require.WithinDuration(t, now.Add(time.Millisecond*100), time.Now(), time.Millisecond*20)
}

func Test_TrippleCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Millisecond * 100)
		cancel()
	}()
	now := time.Now()
	var exitTime1 time.Time
	var exitTime2 time.Time
	var exitTime3 time.Time
	err := Stack(ctx,
		func(ctx context.Context) error {
			<-ctx.Done()
			exitTime1 = time.Now()
			return nil
		},
		func(ctx context.Context) error {
			<-ctx.Done()
			exitTime2 = time.Now()
			return nil
		},
		func(ctx context.Context) error {
			<-ctx.Done()
			exitTime3 = time.Now()
			return errors.New("tripple_cancel")
		},
	)
	require.EqualError(t, err, "tripple_cancel")
	// Default time order: 3 -> 2 -> 1.
	require.True(t, exitTime3.Before(exitTime2))
	require.True(t, exitTime2.Before(exitTime1))
	require.WithinDuration(t, now.Add(time.Millisecond*100), time.Now(), time.Millisecond*20)
}

func Test_TrippleCancel_FirstFailed(t *testing.T) {
	now := time.Now()
	var exitTime1 time.Time
	var exitTime2 time.Time
	var exitTime3 time.Time
	err := Stack(context.Background(),
		func(ctx context.Context) error {
			time.Sleep(time.Millisecond * 100)
			exitTime1 = time.Now()
			return errors.New("tripple_cancel")
		},
		func(ctx context.Context) error {
			<-ctx.Done()
			exitTime2 = time.Now()
			return nil
		},
		func(ctx context.Context) error {
			<-ctx.Done()
			exitTime3 = time.Now()
			return nil
		},
	)
	require.EqualError(t, err, "tripple_cancel")
	// Default time order with time1 as an exception: 1 -> 3 -> 2.
	require.True(t, exitTime3.Before(exitTime2))
	require.True(t, exitTime1.Before(exitTime3))
	require.WithinDuration(t, now.Add(time.Millisecond*100), time.Now(), time.Millisecond*20)
}

func Test_TrippleCancel_SecondFailed(t *testing.T) {
	now := time.Now()
	var exitTime1 time.Time
	var exitTime2 time.Time
	var exitTime3 time.Time
	err := Stack(context.Background(),
		func(ctx context.Context) error {
			<-ctx.Done()
			exitTime1 = time.Now()
			return nil
		},
		func(ctx context.Context) error {
			time.Sleep(time.Millisecond * 100)
			exitTime2 = time.Now()
			return errors.New("tripple_cancel")
		},
		func(ctx context.Context) error {
			<-ctx.Done()
			exitTime3 = time.Now()
			return nil
		},
	)
	require.EqualError(t, err, "tripple_cancel")
	// Default time order with time2 as an exception: 2 -> 3 -> 1.
	require.True(t, exitTime3.Before(exitTime1))
	require.True(t, exitTime2.Before(exitTime3))
	require.WithinDuration(t, now.Add(time.Millisecond*100), time.Now(), time.Millisecond*20)
}

// Test_DoubleCheckContext checks that context values are accessible in
// stacked functions.
func Test_DoubleCheckContext(t *testing.T) {
	type key int
	const cKey key = 0
	golden := 42
	ctx := context.WithValue(context.Background(), cKey, &golden)

	err := Stack(ctx,
		func(ctx context.Context) error {
			require.NotEmpty(t, ctx.Value(cKey))
			return nil
		},
		func(ctx context.Context) error {
			require.NotEmpty(t, ctx.Value(cKey))
			return nil
		},
	)
	require.NoError(t, err)
}
