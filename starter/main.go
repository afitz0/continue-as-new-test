package main

import (
	"context"
	"flag"
	"os"
	"os/exec"
	"time"

	"go.temporal.io/sdk/client"

	"go.uber.org/zap/zapcore"

	"starter"
	"starter/zapadapter"
)

func main() {
	var testFlag string
	flag.StringVar(&testFlag, "test", "all", "Name of test to run, default \"all\". See workflow.go for names.")
	flag.Parse()

	logger := zapadapter.NewZapAdapter(zapadapter.NewZapLogger(zapcore.DebugLevel))
	c, err := client.NewLazyClient(client.Options{
		Logger: logger,
	})
	if err != nil {
		logger.Error("Unable to create client", err)
		os.Exit(1)
	}
	defer c.Close()

	var testEnd starter.TestIdenfier
	var testStart starter.TestIdenfier
	if testFlag == "all" {
		testStart = starter.TEST_NULL_START + 1
		testEnd = starter.TEST_NULL_END - 1
	} else {
		testStart = starter.TestNameToId(testFlag)
		testEnd = testStart
	}

	for i := testStart; i <= testEnd; i++ {
		test := i
		wId := "continue-as-new-test-" + starter.TestIdToName(test)
		workflowOptions := client.StartWorkflowOptions{
			ID:        wId,
			TaskQueue: "can-test-queue",
		}

		var we client.WorkflowRun
		var err error

		if test == starter.SIGNAL_WITH_START {
			we, err = c.SignalWithStartWorkflow(context.Background(), wId, "signal", "", workflowOptions, starter.Workflow, test)
		} else {
			we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, starter.Workflow, test)
		}

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
			// All of these tests should be able to at least start successfully. Fatal if they don't.
			logger.Error("Unable to execute workflow", err)
			os.Exit(1)
		}

		logger.Info("Started workflow", "Test", starter.TestIdToName(test), "WorkflowID", we.GetID(), "RunID", we.GetRunID())
		logger.Debug("Awaiting workflow completion...")
		err = we.Get(context.Background(), nil)
		if err != nil {
			logger.Error("Unable to get workflow results; depending on the test this may be expected (e.g., Terminated due to exceeding history limits)", "error", err)
		}
		logger.Debug("Workflow completed; now retrieving metadata")

		runId := we.GetRunID()
		desc, err := c.DescribeWorkflowExecution(context.Background(), wId, runId)
		wInfo := desc.GetWorkflowExecutionInfo()
		histLength := wInfo.GetHistoryLength()
		histSize := wInfo.GetHistorySizeBytes()

		histSizeFromCli, err := getHistorySize(wId, runId)
		if err != nil {
			logger.Error("Could not get history size from CLI", "Workflow ID", wId, "Run ID", runId)
		}

		logger.Info("Workflow finished",
			"test", starter.TestIdToName(test),
			"history length", histLength,
			"history size (from API)", histSize,
			"history size (downloaded from CLI)", histSizeFromCli)
	}
}

func getHistorySize(wId string, runId string) (int, error) {
	cmd := exec.Command("temporal", "workflow", "show",
		"--output", "json",
		"--no-pager",
		"--fields", "long",
		"--workflow-id", wId,
		"--run-id", runId,
		"--max-field-length", "1000000")
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	return len(out), nil
}
