package starter

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func Workflow(ctx workflow.Context, test Test) error {
	logger := workflow.GetLogger(ctx)

	if test == TEST_QUERY {
		queryType := "query"
		err := workflow.SetQueryHandler(ctx, queryType, func() (string, error) {
			logger.Debug("Received query request")
			return "", nil
		})
		if err != nil {
			logger.Error("failed to register query handler")
			return err
		}
	}

	logger.Info("Workflow started")

	selector := workflow.NewSelector(ctx)
	if test == TEST_ONE_SIGNAL || test == TEST_ENDLESS_SIGNALS {
		// Register signal handler
		signalChannel := workflow.GetSignalChannel(ctx, "signal")

		selector.AddReceive(signalChannel, func(c workflow.ReceiveChannel, _ bool) {
			var signal interface{}
			c.Receive(ctx, &signal)
			logger.Info("Signal received")
		})
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 1.0,
			MaximumInterval:    10 * time.Second,
			MaximumAttempts:    0, // 0 is infinite
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	var a *Activities

	info := workflow.GetInfo(ctx)

	iterations := 1
	if test == TEST_ZERO_SIZE_ACTIVITY || test == TEST_BIG_ACTIVITY {
		iterations = 50 * 1024
	}
	for i := 0; i < iterations; i++ {
		var err error
		switch test {
		case TEST_ZERO_SIZE_ACTIVITY:
			// (near) 0-sized activity. Expected that approximately 8,530 of these activities runs.
			// Each activity execution generates 6 events: 3 for the workflow (scheduled, started,
			// completed) and 3 for the activity (same).
			err = workflow.ExecuteActivity(ctx, a.NilActivity).Get(ctx, nil)
		case TEST_BIG_ACTIVITY:
			// Configurably-sized activities. To hit the 50MB size limit well before the 50K length
			// limit, if each activity returns 500KB of data, this workflow should terminate after
			// ~100 activity executions.
			err = workflow.ExecuteActivity(ctx, a.LargeReturnActivity, int(0.5*1024*1024)).Get(ctx, nil)
		case TEST_TIMER:
			//err = workflow.Sleep(ctx, time.Duration(time.Second*1))
			err = workflow.NewTimer(ctx, time.Duration(time.Second*1)).Get(ctx, nil)
		case TEST_QUERY:
			fallthrough
		case TEST_NO_ACTIVITY:
			fallthrough
		default:
			break
		}

		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return err
		}

		logger.Debug("Completed Activity", "number", i)
		logger.Debug("Workflow event history size", "event history length", info.GetCurrentHistoryLength())
	}

	// Block as necessary for signals
	if test == TEST_ONE_SIGNAL {
		selector.Select(ctx)
	} else if test == TEST_ENDLESS_SIGNALS {
		// Purposefully infinite so that we can trigger the history limit termination
		for {
			selector.Select(ctx)
		}
	}

	logger.Info("Workflow completed.")
	return nil
}
