# Continue-As-New Testing

A project for testing various scenarios for how to determine when Continue-As-New is necessary for
Temporal Workflows. That is, Continue-As-New is necessary to avoid having a Workflow be terminated
for reaching either 50K (51,200) Events or a total Event History Size of 50MB.

Question is: what are the scenarios that would lead to hitting those limits, and how do I measure
or estimate those quantities for my own Workflows?

The currently implemented scenarios are:

* **"no-activity"**: For a empty Workflow that just starts and then completes successfully, how
  many Events does it generate?
* **"zero-size-activity"**: For a Workflow containing nothing but an empty-arg, empty-return
  Activity in a tight loop, how many times can it run before hitting the history *length* limit?
* **"big-activity"**: For a Workflow containing an Activity with configurable-sized return, how
  many times can it run before hitting the history *size* limit?
* **"timer"**: How many events does a Workflow with a single Timer generate?
* **"signal"**: How many events does a Workflow that blocks waiting for a single Signal generate?
* **"signal_with_start"**: How many events does a Workflow that's started with "SignalWithStart"
  generate?
* **"endless-signals"**: For a Workflow that loops infinitely handling as many Signals as come in,
  as fast as it can, how many can it handle before hitting the history *length* limit?
* **"query"**: Do Queries generate any Events?

All of the above presume that everything happens successfully: no failed Activities and no failed
Workflows beyond the expected "Terminated Due to Exceeding Limits."

## Usage

1. Start the worker: `go run worker/main.go`
2. Start the tests: `go run starter/main.go [--test <test name>]`
  * The optional parameter `--test` accepts any one of the test name strings above (see
    `./shared.go` for the real list). Default: "all"
3. Results are logged by the starter program at the "INFO" level.

## Notes

* When I first wrote this, I was running on version 0.5.0 of the [Temporal CLI](https://github.com/temporalio/cli).
  In that version (woefully outdated), while the DescribeWorkflowExecution -> GetWorkflowExecutionInfo
  -> GetHistorySizeBytes API existed, I found that it always returned 0. After updating to 0.8.0, it
  actually returns a real value.

## TODOs

* Many of the above questions aren't directly answered. Add counter outputs for things like "how
  many activites were actually run"
* Make `--test` accept CSV

