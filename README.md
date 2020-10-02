# goasciidoc
Document your go code using [asciidoc](http://asciidoctor.org/). It allows you to have asciidoc [markup](https://asciidoctor.org/docs/asciidoc-writers-guide/) 
in all code documentation. Asciidoc do support many plugins to e.g. render sequence diagrams, svg images, ERD, BPMN, RackDiag and many more, one such component is
[kroki](https://kroki.io/#support) that renders the ascii into fineart :).

To generate documentation for this project as mydoc.adoc, do the following:
```bash
goasciidoc -o mydoc.adoc -c "{\"toc\":\"Table of Contents\"}"
```

The above will generate internal, test and provides a set of overrides to the index (first section of the document). No templates where overridden but it is quite possible to render however you like. Take a look at `defaults.go` to view how such may look like. It is standard go templates.

You may have more properties in the `-c` (configuration) parameter, for example:
```json
 {
  "author": "Mario Toffia",
  "email": "mario.toffia@xy.net",
  "web": "https://github.com/mariotoffia/goasciidoc",
  "images": "../meta/assets",
  "title": "Go Asciidoc Document Generator",
  "toc": "Table of Contents",
  "toclevel": 2
}
```

## Thanks
The package `goparser` was taken from an open source project [by zpatrick](https://github.com/zpatrick/go-parser). It seemed abandoned so I've integrated it into this project (and extended it) and now it deviates rather much from it's earlier pure form ;). Many thanks @zpatrick!! That part has a [MIT License](https://github.com/zpatrick/go-parser/blob/master/LICENSE).

`copy.go` is created by Roland Singer [roland.singer@desertbit.com] and is used for unit test. Many thanks @r0l1. You may find the original [here](https://gist.github.com/r0l1/92462b38df26839a3ca324697c8cba04).

## Install
1) `git clone https://github.com/mariotoffia/goasciidoc.git`
2) `cd goasciidoc; go install`
3) `goasciidoc --stdout` (documentation should be rendered on stdout)

If you encounter any problems please check the following:

1) make sure your `$GOPATH/bin` in your path `export PATH=$PATH:$GOPATH/bin`
2) Make sure your environment is properly setup, my environment WSL2 (ubuntu 18.02) is
  ```bash
   export GOOS=linux
   export GOARCH=amd64
   export GO111MODULE=on
  ```

You may now use the `goasciidoc` e.g. in the `goasciidoc` repo by `goasciidoc --stdout`. This will emit this project documentation onto the stdout. If you need
help on flags and parameters jus do a `goasciidoc --help` to display (**note this may be old output**):

```bash
goasciidoc v0.0.3
Usage: goasciidoc [--out PATH] [--stdout] [--module PATH] [--internal] [--private] [--test] [--noindex] [--indexconfig JSON] [--overrides OVERRIDES] [PATH [PATH ...]]

Positional arguments:
  PATH                   Directory or files to be included in scan (if none, current path is used)

Options:
  --out PATH, -o PATH    The out filepath to write the generated document, default module path, file docs.adoc
  --stdout               If output the generated asciidoc to stdout instead of file
  --module PATH, -m PATH
                         an optional folder or file path to module, otherwise current directory
  --internal, -i         If internal go code shall be rendered as well
  --private, -p          If files beneath directories starting with an underscore shall be included
  --test, -t             If test code should be included
  --noindex, -n          If no index header shall be genereated
  --indexconfig JSON, -c JSON
                         JSON document to override the IndexConfig
  --overrides OVERRIDES, -t OVERRIDES
                         name=template filepath to override default templates
  --help, -h             display this help and exit
  --version              display version and exit
```
## Notes
This project consists of a parser to parse go-code and a producer to produce asciidoc files from the code & code documentation. It bases its rendering system heavily on templates (`asciidoc/template.go`) with some "sane" default so it may be rather easily overridden.
