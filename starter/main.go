package main

import (
	"context"
	"os"
	"time"

	"go.temporal.io/sdk/client"

	"go.uber.org/zap/zapcore"

	"starter"
	"starter/zapadapter"
)

func main() {
	logger := zapadapter.NewZapAdapter(zapadapter.NewZapLogger(zapcore.DebugLevel))
	c, err := client.NewLazyClient(client.Options{
		Logger: logger,
	})
	if err != nil {
		logger.Error("Unable to create client", err)
		os.Exit(1)
	}
	defer c.Close()

	for i := starter.TEST_NULL_START + 1; i < starter.TEST_NULL_END; i++ {
		test := i
		wId := "continue-as-new-test-" + starter.GetTestName(test)
		workflowOptions := client.StartWorkflowOptions{
			ID:        wId,
			TaskQueue: "can-test-queue",
		}
		we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, starter.Workflow, test)
		switch test {
		case starter.SIGNAL:
			// Give enough time for the Workflow to start then yield back to the server.
			time.Sleep(time.Duration(time.Second * 5))
			err = c.SignalWorkflow(context.Background(), wId, we.GetRunID(), "signal", "")
			if err != nil {
				logger.Error("Unable to signal workflow", err)
				os.Exit(1)
			}
		case starter.QUERY:
			_, err := c.QueryWorkflow(context.Background(), wId, we.GetRunID(), "query", "")
			if err != nil {
				logger.Error("Unable to query workflow", err)
				os.Exit(1)
			}
		}
		if err != nil {
			logger.Error("Unable to execute workflow", err)
			os.Exit(1)
		}

		logger.Info("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
		logger.Info("Awaiting workflow completion...")
		err = we.Get(context.Background(), nil)
		if err != nil {
			logger.Error("Unable to get workflow results; depending on the test this may be expected (e.g., Terminated due to exceeding history limits)", err)
		}

		desc, err := c.DescribeWorkflowExecution(context.Background(), wId, we.GetRunID())
		histLength := desc.WorkflowExecutionInfo.GetHistoryLength()
		histSize := desc.WorkflowExecutionInfo.GetHistorySizeBytes()

		logger.Info("Workflow finished",
			"test", starter.GetTestName(test),
			"history length", histLength,
			"history size", histSize)
	}
}
