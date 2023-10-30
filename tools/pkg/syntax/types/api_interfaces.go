package types

func (r *KformBlockAttributes) GetSource() string {
	if r != nil && r.Source != nil {
		return *r.Source
	}
	return ""
}