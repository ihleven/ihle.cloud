package llist_test

import (
	"fmt"

	"github.com/ihleven/ihlvn/pkg/llist"
)

// The demo that arrived as a main function inside the package, unchanged.
//
// It is kept because its output is the interesting part. Add inserts after the
// node it is called on, and every call here is on the head, so each number
// lands in second place and the list reads back in the reverse of the order it
// was built — all but the 1, which is the head and never moves.
func Example() {
	list := llist.NewList(1)
	list.Add(2)
	list.Add(6)
	list.Add(345)
	list.Add(4)
	list.Add(782)
	fmt.Printf("Hello world! [ %v \n", list)
	// Output:
	// Hello world! [ 1, 782, 4, 345, 6, 2 ]
}
