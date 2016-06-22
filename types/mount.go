package types

// Mount type is used for mounting and storing state
type Mount struct {
	UUID               string
	Source, Target     string
	Upper, Work, Merge string
}
