// +build todo

package ini_test

import (
	"fmt"
	"github.com/DeedleFake/ini"
	"strings"
)

func ExampleNewDecoder() {
	const example = `[Section]
this=is
an=example`

	d := ini.NewDecoder(strings.NewReader(example))

	m := make(map[string]map[string]string)
	err := d.Decode(m)
	if err != nil {
		panic(err)
	}

	fmt.Printf("this: %v\n", m["Section/this"])
	fmt.Printf("an: %v\n", m["Section/an"])
	// Output: this: is
	// an: example
}
