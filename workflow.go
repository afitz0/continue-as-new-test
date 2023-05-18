package continue_as_new_test

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func Workflow(ctx workflow.Context, test Test) (err error) {
	logger := workflow.GetLogger(ctx)

	if test == TestQuery {
		queryType := "query"
		err = workflow.SetQueryHandler(ctx, queryType, func() (string, error) {
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
	if test == TestOneSignal || test == TestEndlessSignals {
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
	if test == TestZeroSizeActivity || test == TestBigActivity {
		iterations = 50 * 1024
	} else if test == TestCANAbandonedActivities {
		iterations = 10
		if info.ContinuedExecutionRunID != "" {
			// In the continue workflow, don't start any new activities.
			iterations = 0
		}
	}

	for i := 0; i < iterations; i++ {
		var err error
		switch test {
		case TestZeroSizeActivity:
			// (near) 0-sized activity. Expected that approximately 8,530 of these activities runs.
			// Each activity execution generates 6 events: 3 for the workflow (scheduled, started,
			// completed) and 3 for the activity (same).
			err = workflow.ExecuteActivity(ctx, a.NilActivity).Get(ctx, nil)
		case TestBigActivity:
			// Configurably-sized activities. To hit the 50MB size limit well before the 50K length
			// limit, if each activity returns 500KB of data, this workflow should terminate after
			// ~100 activity executions.
			err = workflow.ExecuteActivity(ctx, a.LargeReturnActivity, int(0.5*1024*1024)).Get(ctx, nil)
		case TestTimer:
			//err = workflow.Sleep(ctx, time.Duration(time.Second*1))
			err = workflow.NewTimer(ctx, time.Duration(time.Second*1)).Get(ctx, nil)
		case TestCANAbandonedActivities:
			// Async start the activity, will continue-as-new later
			_ = workflow.ExecuteActivity(ctx, a.AsyncActivity, i)
			// Sleeping between activities changes the behavior of this test.
			//workflow.Sleep(ctx, time.Duration(time.Second*1))
		case TestQuery:
			fallthrough
		case TestGoActivity:
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
	if test == TestOneSignal {
		selector.Select(ctx)
	} else if test == TestEndlessSignals {
		// Purposefully infinite so that we can trigger the history limit termination
		for {
			selector.Select(ctx)
		}
	}

	// Continue-As-New, if the test requires it.
	if test == TestCANAbandonedActivities {
		// For this test, only allow CAN once.
		if info.ContinuedExecutionRunID == "" {
			return workflow.NewContinueAsNewError(ctx, Workflow, test)
		}
	}

	logger.Info("Workflow completed.")
	return nil
}
