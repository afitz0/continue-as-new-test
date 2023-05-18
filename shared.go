package continue_as_new

const TaskQueueName = "can-test-queue"

type Test string

func (t Test) GetId() int {
	for i, value := range Tests {
		if value == t {
			return i
		}
	}
	return -1
}

const (
	TestZeroSizeActivity       Test = "zero-size-activity"
	TestBigActivity            Test = "big-activity"
	TestGoActivity             Test = "no-activity"
	TestTimer                  Test = "timer"
	TestOneSignal              Test = "signal"
	TestSignalWithStart        Test = "signal-with-start"
	TestEndlessSignals         Test = "endless-signals"
	TestQuery                  Test = "query"
	TestCANAbandonedActivities Test = "abandoned-activities"
)

var Tests = []Test{
	TestZeroSizeActivity,
	TestBigActivity,
	TestGoActivity,
	TestTimer,
	TestOneSignal,
	TestSignalWithStart,
	TestEndlessSignals,
	TestQuery,
	TestCANAbandonedActivities,
}
