package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"go.uber.org/zap/zapcore"

	can_test "github.com/afitz0/continue-as-new-test"
	"github.com/afitz0/continue-as-new-test/zapadapter"
)

func main() {
	c, err := client.NewLazyClient(client.Options{
		Logger: zapadapter.NewZapAdapter(
			zapadapter.NewZapLogger(zapcore.WarnLevel)),
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "can-test-queue", worker.Options{})

	a := &can_test.Activities{}
	w.RegisterWorkflow(can_test.Workflow)
	w.RegisterActivity(a)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
