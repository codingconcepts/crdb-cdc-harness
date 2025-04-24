package models

// StringFlags allows the application to take multiple arguments with the
// same name and combine them into a slice.
type StringFlags []string

func (i *StringFlags) String() string {
	return "my string representation"
}

func (i *StringFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
