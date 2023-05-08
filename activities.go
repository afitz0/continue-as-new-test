package starter

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
