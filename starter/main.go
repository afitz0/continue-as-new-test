package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"go.temporal.io/sdk/client"

	"go.uber.org/zap/zapcore"

	can_test "github.com/afitz0/continue-as-new-test"
	"github.com/afitz0/continue-as-new-test/zapadapter"
)

func main() {
	var testFlag string
	flag.StringVar(&testFlag, "test", "all", "Name of test to run, default \"all\". See workflow.go for names.")
	flag.Parse()

	logger := zapadapter.NewZapAdapter(zapadapter.NewZapLogger(zapcore.DebugLevel))
	c, err := client.Dial(client.Options{
		Logger: logger,
	})
	if err != nil {
		logger.Error("Unable to create client", err)
		os.Exit(1)
	}
	defer c.Close()

	requestedTest := can_test.Test{Name: testFlag}
	if testFlag != "all" && requestedTest.GetId() == -1 {
		logger.Error("Unknown test name")
		os.Exit(1)
	}

	testStart := 0
	testEnd := len(can_test.Tests)
	if testFlag != "all" {
		testStart = requestedTest.GetId()
		testEnd = testStart + 1
	}

	for i := testStart; i < testEnd; i++ {
		test := can_test.Tests[i]
		wId := "continue-as-new-test-" + requestedTest.GetName()
		workflowOptions := client.StartWorkflowOptions{
			ID:        wId,
			TaskQueue: "can-test-queue",
		}

		var we client.WorkflowRun
		var err error

		if test == can_test.TEST_SIGNAL_WITH_START {
			we, err = c.SignalWithStartWorkflow(context.Background(), wId, "signal", "", workflowOptions, can_test.Workflow, test)
		} else {
			we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, can_test.Workflow, test)
		}

		switch test {
		case can_test.TEST_ONE_SIGNAL:
			// Give enough time for the Workflow to start then yield back to the server.
			time.Sleep(time.Duration(time.Second * 5))
			err = c.SignalWorkflow(context.Background(), wId, we.GetRunID(), "signal", "")
			if err != nil {
				logger.Error("Unable to signal workflow", "error", err)
				os.Exit(1)
			}
		case can_test.TEST_ENDLESS_SIGNALS:
			// Queue up 51K signals, understanding that many will be dropped.
			for i := 0; i < 51*1024; i++ {
				err = c.SignalWorkflow(context.Background(), wId, we.GetRunID(), "signal", fmt.Sprint(i))
				if err != nil {
					logger.Warn("Unable to signal workflow; this is expected", "error", err)
					err = nil
					break
				}
			}
		case can_test.TEST_QUERY:
			_, err := c.QueryWorkflow(context.Background(), wId, we.GetRunID(), "query", "")
			if err != nil {
				logger.Error("Unable to query workflow", "error", err)
				os.Exit(1)
			}
		}

		if err != nil {
			// All of these tests should be able to at least start successfully. Fatal if they don't.
			logger.Error("Unable to execute workflow", "error", err)
			os.Exit(1)
		}

		logger.Info("Started workflow", "Test", test.GetName(), "WorkflowID", we.GetID(), "RunID", we.GetRunID())
		logger.Debug("Awaiting workflow completion...")
		err = we.Get(context.Background(), nil)
		if err != nil {
			logger.Warn("Unable to get workflow results; this is probably expected (e.g., Terminated due to exceeding history limits)", "error", err)
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
			"test", test.GetName(),
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
