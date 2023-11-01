package vars

const DummyKey = "BamBoozle"

type Variable struct {
	// Data contains the result of the block.
	// For module blockType output we can have multiple entries, so we store them using a key in the map
	// For all other blockTypes we use a dummy key
	Data map[string][]any
}
