package js

type dateObject struct {
}

func (d *dateObject) GetTimezoneOffset(argument []interface{}) interface{} {
	return 0
}

// Date simulates Date function
func Date(argument []interface{}) interface{} {
	return new(dateObject)
}
