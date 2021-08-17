//nolint:paralleltest
package web_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	api "go.xsfx.dev/schnutibox/pkg/api/v1"
	"go.xsfx.dev/schnutibox/pkg/web"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestTimerService(t *testing.T) {
	tables := []struct {
		name     string
		req      *api.Timer
		expected *api.Timer
		err      error
	}{
		{
			"10 seconds",
			&api.Timer{Duration: &durationpb.Duration{Seconds: 10}},
			&api.Timer{Duration: &durationpb.Duration{Seconds: 10}},
			nil,
		},
	}

	for _, table := range tables {
		table := table
		t.Run(table.name, func(t *testing.T) {
			t.Parallel()
			require := require.New(t)
			ctx := context.Background()
			timerSvc := web.TimerServer{}
			resp, err := timerSvc.Create(ctx, &api.Timer{Duration: &durationpb.Duration{Seconds: 10}})
			if table.err == nil {
				require.NoError(err)
			}

			require.Equal(table.expected, resp)
		})
	}
}
