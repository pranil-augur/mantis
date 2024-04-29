package datamodel

import (
	"github.com/opentofu/opentofu/internal/hof/schema/dm"
)

MyObject: dm.Object & {

	foo: "bar"
	ans: 42

	animals: {
		cow: "moo"
		cat: "meow"
		dog: "woof"
	}
}
