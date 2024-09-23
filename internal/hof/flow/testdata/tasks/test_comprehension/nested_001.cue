package test

import (
	"strings"
)

#censusData: [
	{name: "Kinshasa", pop: 16_315_534},
	{name: "Lagos", pop: 15_300_000},
	{name: "Cairo", pop: 10_100_166},
	{name: "Giza", pop: 9_250_791},
]


@flow()
nested: {
	tasks: {
		get: {
			FOO: {
				for index, value in #censusData
				let lower = strings.ToLower(value.name) {
					"\(lower)": {
						population: value.pop
						name:       value.name
						position:   index + 1
					}
				}
			}
		} 
	}
}

 

