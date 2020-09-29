package main

import (
	"fmt"
	"log"

	"github.com/mariotoffia/goasciidoc/goparser"
)

func main() {
	src := `
// The package foo is a sample package.
package foo

import (
	"fmt"
	"time"
)
// bar is a method to print the time.
//
// Example:
// bar()
func bar() {
	fmt.Println(time.Now())
}`

	f, err := goparser.ParseInlineFile(src)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%#v\n", f)

	for _, s := range f.StructMethods {
		fmt.Printf("%#v\n", s)
	}
}
