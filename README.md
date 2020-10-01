# goasciidoc
Document your go code using asciidoc instead of godoc. This project consists of a parser to parse go-code and a producer to produce asciidoc files from the code & code documentation. It bases its rendering system heavily on templates (`asciidoc/template.go`) with some "sane" default so it may be rather easily overridden.

## Thanks
The package `goparser` was taken from an open source project [by zpatrick](https://github.com/zpatrick/go-parser). It seemed abandoned so I've integrated it into this project (and extended it) and now it deviates rather much from it's earlier pure form ;). Many thanks @zpatrick!! That part has a [MIT License](https://github.com/zpatrick/go-parser/blob/master/LICENSE).
