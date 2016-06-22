package types

type Mount struct {
	UUID               string
	Source, Target     string
	Upper, Work, Merge string
}
