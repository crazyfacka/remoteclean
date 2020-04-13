package domain

import "time"

// Item represents a deletable item
type Item struct {
	Created  time.Time
	FullPath string
}

// Items represents a set of items
type Items []Item

// Len implements the Len from the sort interface
func (its Items) Len() int {
	return len(its)
}

// Swap implements the Swap from the sort interface
func (its Items) Swap(i, j int) {
	its[i], its[j] = its[j], its[i]
}

// Less implements the Less from the sort interface
func (its Items) Less(i, j int) bool {
	return its[i].Created.Before(its[j].Created)
}
