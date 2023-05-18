package continue_as_new_test

const TaskQueueName = "can-test-queue"

type Test struct {
	Name string
}

func (t Test) GetName() string {
	return t.Name
}

func (t Test) GetId() int {
	for i, value := range Tests {
		if value == t {
			return i
		}
	}
	return -1
}

var (
	TestZeroSizeActivity       = Test{"zero-size-activity"}
	TestBigActivity            = Test{"big-activity"}
	TestGoActivity             = Test{"no-activity"}
	TestTimer                  = Test{"timer"}
	TestOneSignal              = Test{"signal"}
	TestSignalWithStart        = Test{"signal-with-start"}
	TestEndlessSignals         = Test{"endless-signals"}
	TestQuery                  = Test{"query"}
	TestCANAbandonedActivities = Test{"abandoned-activities"}
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
