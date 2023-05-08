package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/client"

	"starter"
	"starter/zapadapter"
)

func main() {
	c, err := client.NewLazyClient(client.Options{
		Logger: zapadapter.NewZapAdapter(
			zapadapter.NewZapLogger()),
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	wId := "temporal-starter-workflow"
	workflowOptions := client.StartWorkflowOptions{
		ID:        wId,
		TaskQueue: "temporal-starter",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, starter.Workflow)
	switch starter.TEST {
	case starter.SIGNAL:
		// Give enough time for the Workflow to start then yield back to the server.
		time.Sleep(time.Duration(time.Second * 5))
		err = c.SignalWorkflow(context.Background(), wId, we.GetRunID(), "signal", "")
		if err != nil {
			log.Fatalln("Unable to signal workflow", err)
		}
	case starter.QUERY:
		_, err := c.QueryWorkflow(context.Background(), wId, we.GetRunID(), "query", "")
		if err != nil {
			log.Fatalln("Unable to query workflow", err)
		}
	default:
		we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, starter.Workflow)
	}
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
	log.Println("Awaiting workflow completion...")
	err = we.Get(context.Background(), nil)
	if err != nil {
		log.Fatalln("Unable to get workflow results", err)
	}

	desc, err := c.DescribeWorkflowExecution(context.Background(), wId, we.GetRunID())
	histLength := desc.WorkflowExecutionInfo.GetHistoryLength()
	histSize := desc.WorkflowExecutionInfo.GetHistorySizeBytes()

	log.Println(fmt.Sprintf("Workflow running test id %v finished with history length (%v) and size (%v bytes)", starter.TEST, histLength, histSize))
}
