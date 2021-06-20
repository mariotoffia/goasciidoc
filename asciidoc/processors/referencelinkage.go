package processors

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ahmetb/go-linq/v3"
	"github.com/mariotoffia/goasciidoc/goparser"
)

// ReferenceLinkerProcessor will substitute any found reference within
// backtick tags. By default no prefix is required, but may be set in
// the _prefix_ in order to be activated.
//
// This will search the package and type in imports and in current package
// to render a cross reference.
//
// .Allowable Reference Types
// |===
// |Expression |Type
//
// |MyRef 					|Same package ref
// |package.MyRef 			|Other package ref
// |MyRef.MyRef2 			|Struct Field, Interface Function in same package
// |MyRef(MyRef2) 			|Receiver function
// |package.MyRef.MyRef2 	|Struct Field, Interface Function in other package
// |package.MyRef(MyRef2) 	|Receiver function in other package
// |===
//
// The _MyRef_ is resolved in the following way
//
// 1. Interface
// 2. Struct
// 3. Variable
// 4. TypeDef Variable
// 5. TypeDef Function
// 6. Plain method
// 7. Receiver, and have only one receiver
//
// If multiple receiver for same function name, use paranthesis and the name of the receiver
// to address. For example, _MyFunc(*MyStruct)_.
func ReferenceLinkerProcessor(prefix string) DocProcessor {

	base := `([a-zA-Z0-9\.]+)([\(][a-zA-Z0-9*]*[\)])?`
	r, _ := regexp.Compile("`" + base + "`")

	// https://github.com/golang/example/tree/master/gotypes#introduction
	// https://github.com/golang/go/issues/11415#issuecomment-283445198
	// https://github.com/golang/tools

	return func(module *goparser.GoModule, file *goparser.GoFile, doc, fq string) string {

		for _, match := range r.FindAllString(doc, -1) {

			idx := strings.Index(match, "(")
			receiver := ""
			expr := match

			if idx != -1 {

				receiver = match[idx:strings.Index(match, ")")]
				expr = match[0:idx]
			}

			ref := expr[strings.LastIndex(expr, "."):]

			components := strings.Split(expr[0:strings.Index(expr, ".")], ".")

			if len(components) == 2 {
				// Need to checkup imports of packages to
				// determine if package or ref
				linq.From(file.Imports).
					Where(func(i interface{}) bool {
						imp := i.(*goparser.GoImport)
						// TODO: Not all modules ending do not have the same package
						// TODO: e.g. see linq import and linq package differ!
						return imp.Name == components[0] ||
							imp.TryResolvePackageName() == components[0]
					})
			}

			fmt.Println(components)
			fmt.Println(ref)
			fmt.Println(receiver)

			// TODO: resolve the package (if any) and type, method, interface etc...
		}

		return doc
	}

}
