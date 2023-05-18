package main

import (
	"flag"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"go.uber.org/zap/zapcore"

	can_test "github.com/afitz0/continue-as-new-test"
	"github.com/afitz0/continue-as-new-test/zapadapter"
)

func main() {
	var workerCount int
	flag.IntVar(&workerCount, "n", 1, "How many worker threads to start. Default: 1")
	flag.Parse()

	c, err := client.Dial(client.Options{
		Logger: zapadapter.NewZapAdapter(
			zapadapter.NewZapLogger(zapcore.WarnLevel)),
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	var pool []worker.Worker

	log.Println("Starting", workerCount, "workers")
	for i := 0; i < workerCount; i++ {
		w := worker.New(c, can_test.TaskQueueName, worker.Options{
			Identity: fmt.Sprintf("worker-%v", i),
		})

		a := &can_test.Activities{}
		w.RegisterWorkflow(can_test.Workflow)
		w.RegisterActivity(a)

		go func() {
			err := w.Run(nil)
			if err != nil {
				log.Fatalln("Unable to start worker", err)
			}
			pool = append(pool, w)
		}()
	}

	<-worker.InterruptCh()
	for _, w := range pool {
		w.Stop()
	}
}
