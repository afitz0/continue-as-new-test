package continue_as_new_test

const TASK_QUEUE_NAME = "can-test-queue"

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
	TEST_ZERO_SIZE_ACTIVITY       = Test{"zero-size-activity"}
	TEST_BIG_ACTIVITY             = Test{"big-activity"}
	TEST_NO_ACTIVITY              = Test{"no-activity"}
	TEST_TIMER                    = Test{"timer"}
	TEST_ONE_SIGNAL               = Test{"signal"}
	TEST_SIGNAL_WITH_START        = Test{"signal-with-start"}
	TEST_ENDLESS_SIGNALS          = Test{"endless-signals"}
	TEST_QUERY                    = Test{"query"}
	TEST_CAN_ABANDONED_ACTIVITIES = Test{"abandoned-activities"}
)

var Tests = [...]Test{
	TEST_ZERO_SIZE_ACTIVITY,
	TEST_BIG_ACTIVITY,
	TEST_NO_ACTIVITY,
	TEST_TIMER,
	TEST_ONE_SIGNAL,
	TEST_SIGNAL_WITH_START,
	TEST_ENDLESS_SIGNALS,
	TEST_QUERY,
	TEST_CAN_ABANDONED_ACTIVITIES,
}
