# llist

A singly-linked list, moved here from `github.com/ihleven/pkg` when that
repository was emptied. It is a sketch rather than a finished thing, and it is
here to be played with, so nothing below has been fixed — this is a list of what
you will run into, in the order you are likely to run into it.

The demo that arrived as a `main` function inside the package is now
`Example` in `linkedlist_test.go`:

    go test ./pkg/llist -run Example -v

Its output is pinned, so any change to the ordering below will fail the test
rather than pass quietly.

## `NewList` returns an unexported type

`NewList` hands back a `*elem`, and `elem` is not exported. Outside the package
you can hold the value — `list := llist.NewList(1)` infers the type — but you
cannot *name* it: no `var` declaration, no struct field, no function parameter,
no return type. This is the first thing that will get in your way, and the
reason `list.go` most likely exists.

## `Add` does not append

It inserts immediately after the node it is called on:

```go
func (e *elem) Add(a interface{}) {
    elem := elem{content: a, next: e.next}
    e.next = &elem
}
```

That is a correct insert-after primitive with a name that promises something
else. In the example every call is on the head, so each new number lands in
second place and the list comes back as `1, 782, 4, 345, 6, 2` — the reverse of
the order it was built, except for the head, which never moves.

Appending would mean walking to the end, which is the operation this shape is
bad at and the usual reason a list keeps a tail pointer.

## `String` prints only half its brackets

It ends the list with `]` but never opens with `[`. The opening bracket in the
example comes from the `Printf` format string, not from the type, so a list
printed any other way looks unterminated at the front.

It is also recursive, one frame per element, and it renders through `%v` on
`e.next` — which works because the value receiver puts `String` in `*elem`'s
method set too.

## `list.go` is an empty struct

```go
type LinkedList struct {
}
```

No fields, no methods, nothing references it. It looks like where a second
attempt was going to go — probably the exported type that `NewList` should be
returning.

## `interface{}`, not a type parameter

The package predates generics. `content interface{}` means every read needs a
type assertion and the compiler will not stop you mixing a string into a list of
ints. `[T any]` is the obvious thing to reach for and would remove the assertions
entirely — worth knowing that the two changes interact: a generic `List[T]`
exported properly is the same edit as fixing the first point above.
