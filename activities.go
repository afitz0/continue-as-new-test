package continue_as_new_test

import "fmt"

type Activities struct{}

func (a *Activities) NilActivity() error {
	return nil
}

func (a *Activities) LargeReturnActivity(bytes int) ([]int, error) {
	data := []int{}

	for i := 0; i < int(bytes/2); i++ {
		data = append(data, 0)
	}
	return data, nil
}

func (a *Activities) AsyncActivity(id int) error {
	fmt.Sprintln("Running async activity; ID is: ", id)
	return nil
}
