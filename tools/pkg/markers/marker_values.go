package markers

// MarkerValues are all the values for some set of markers.
type MarkerValues map[string][]any

// Get fetches the first value that for the given marker, returning
// nil if no values are available.
func (v MarkerValues) Get(name string) any {
	vals := v[name]
	if len(vals) == 0 {
		return nil
	}
	return vals[0]
}
