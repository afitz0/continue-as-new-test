package continue_as_new_test

import "fmt"

type Activities struct{}

func (a *Activities) NilActivity() error {
	return nil
}

func (a *Activities) LargeReturnActivity(count int) ([]int, error) {
	return make([]int, bytes/2), nil
}

func (a *Activities) AsyncActivity(id int) error {
	fmt.Println("Running async activity; ID is: ", id)
	return nil
}
