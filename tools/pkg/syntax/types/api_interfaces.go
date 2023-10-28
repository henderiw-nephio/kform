package types

func (r *KformBlockAttributes) GetSource() string {
	if r.Source != nil {
		return *r.Source
	}
	return ""
}