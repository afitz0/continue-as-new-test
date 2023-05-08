package starter

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type TestIdenfier int

const (
	ZERO_SIZE_ACTIVITY TestIdenfier = iota
	BIG_ACTIVITY
	NO_ACTIVITY
	TIMER
	SIGNAL
	QUERY
)

const TEST = TIMER

// const ITERATIONS = 50 * 1024
const ITERATIONS = 1

func Workflow(ctx workflow.Context) error {
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

	logger := workflow.GetLogger(ctx)
	logger.Info("Workflow started")

	selector := workflow.NewSelector(ctx)
	if TEST == SIGNAL {
		// Register signal handler
		signalChannel := workflow.GetSignalChannel(ctx, "signal")

		selector.AddReceive(signalChannel, func(c workflow.ReceiveChannel, _ bool) {
			var signal interface{}
			c.Receive(ctx, &signal)
			logger.Info("Signal received")
		})
	}

	var a *Activities

	info := workflow.GetInfo(ctx)

	for i := 0; i < ITERATIONS; i++ {
		var err error
		switch TEST {
		case ZERO_SIZE_ACTIVITY:
			// (near) 0-sized activity. Expected that approximately 8,530 of these activities runs.
			// Each activity execution generates 6 events: 3 for the workflow (scheduled, started,
			// completed) and 3 for the activity (same).
			err = workflow.ExecuteActivity(ctx, a.NilActivity).Get(ctx, nil)
		case BIG_ACTIVITY:
			// Configurably-sized activities. To hit the 50MB size limit well before the 50K length
			// limit, if each activity returns 500KB of data, this workflow should terminate after
			// ~100 activity executions.
			err = workflow.ExecuteActivity(ctx, a.LargeReturnActivity, 512*1024).Get(ctx, nil)
		case TIMER:
			//err = workflow.Sleep(ctx, time.Duration(time.Second*1))
			err = workflow.NewTimer(ctx, time.Duration(time.Second*1)).Get(ctx, nil)
		case NO_ACTIVITY:
		default:
			break
		}

		if err != nil {
			logger.Error("Activity failed.", "Error", err)
			return err
		}

		logger.Info("Completed Activity", "number", i)
		logger.Info("Workflow event history size", "event history length", info.GetCurrentHistoryLength())
	}

	if TEST == SIGNAL {
		selector.Select(ctx)
	}

	logger.Info("Workflow completed.")
	return nil
}
