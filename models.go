package main

import (
	"time"

	"gorm.io/gorm"
)

// todo: rename to `clipping`
type NotedItem struct {
	gorm.Model

	// todo: make title + time as primary key
	// todo: make it polymorphic as in sqlalchemy

	// todo: move title and author to separate tables
	// todo: add index for not duplicating notes
	Title  string
	Author string

	// todo: convert to union type
	Type_    string
	Location string
	Time     time.Time

	Highlight string
}
