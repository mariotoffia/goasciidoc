[![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/mod/github.com/mariotoffia/goasciidoc)
[![GitHub Actions](https://img.shields.io/github/workflow/status/mariotoffia/goasciidoc/Go?style=flat-square)](https://github.com/mariotoffia/goasciidoc/actions?query=workflow%3AGo)
![CodeQL](https://github.com/mariotoffia/goasciidoc/workflows/CodeQL/badge.svg)

# goasciidoc
Document your go code using [asciidoc](http://asciidoctor.org/). It allows you to have asciidoc [markup](https://asciidoctor.org/docs/asciidoc-writers-guide/) 
in all code documentation. Asciidoc do support many plugins to e.g. render sequence diagrams, svg images, ERD, BPMN, RackDiag and many more.

One such component is [kroki](https://kroki.io/#support) that renders the ascii into fine-art :).

:bulb: **See the [plugins](#plugins) section below for examples!!**

To generate documentation for this project as mydoc.adoc, do the following:
```bash
goasciidoc -o mydoc.adoc
```

The above will generate standard code documentation, internal and test is excluded. By default it renders a index with some defaults including a table of contents. Is is possible to override the contents by supplying a JSON string with overrides.

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

Everything is rendered using go templates and it is possible to override each of them using the `-t` switch. Take a look at `defaults.go` to view how such may look like. It is standard go templates.

All code is parsed thus you may annotate with asciidoc wherever you want, e.g. 

```golang
// HealthChecker is responsible for doing various health checks on patients.
// 
// Its main flow is conceptualized on following sequence diagram
//
// [mermaid,config-override,svg]
// ....
// sequenceDiagram
//    participant Alice
//    participant Bob
//    Alice->John: Hello John, how are you?
//    loop Healthcheck
//        John->John: Fight against hypochondria
//    end
//    Note right of John: Rational thoughts prevail...
//    John-->Alice: Great!
//    John->Bob: How about you?
//    Bob-->John: Jolly good!
// ....
type HealthChecker struct {
  
}
```

## Install

```bash
go get -u github.com/mariotoffia/goasciidoc
```

You may now use the `goasciidoc` e.g. in the `goasciidoc` repo by `goasciidoc --stdout`. This will emit this project documentation onto the stdout. If you need help on flags and parameters jus do a `goasciidoc --h`.


```bash
goasciidoc v0.2.0
Usage: goasciidoc [--out PATH] [--stdout] [--module PATH] [--internal] [--private] [--nonexported] [--test] [--noindex] [--notoc] [--indexconfig JSON] [--overrides OVERRIDES] [--list-template] [--out-template OUT-TEMPLATE] [--packagedoc FILEPATH] [PATH [PATH ...]]

Positional arguments:
  PATH                   Directory or files to be included in scan (if none, current path is used)

Options:
  --out PATH, -o PATH    The out filepath to write the generated document, default module path, file docs.adoc
  --stdout               If output the generated asciidoc to stdout instead of file
  --module PATH, -m PATH
                         an optional folder or file path to module, otherwise current directory
  --internal, -i         If internal go code shall be rendered as well
  --private, -p          If files beneath directories starting with an underscore shall be included
  --nonexported          Renders Non exported as well as the exported. Default only Exported is rendered.
  --test, -t             If test code should be included
  --noindex, -n          If no index header shall be generated
  --notoc                Removes the table of contents if index document
  --indexconfig JSON, -c JSON
                         JSON document to override the IndexConfig
  --overrides OVERRIDES, -r OVERRIDES
                         name=template filepath to override default templates
  --list-template        Lists all default templates in the binary
  --out-template OUT-TEMPLATE
                         outputs a template to stdout
  --packagedoc FILEPATH, -d FILEPATH
                         set relative package search filepaths for package documentation
  --help, -h             display this help and exit
  --version              display version and exit
```

## Overriding Default Package Overview
By default `goasciidoc` will use _overview.adoc_ or _\_design/overview.adoc_ to generate the package overview. If those are not found, it will default back to the _golang_ package documentation (if any).

It is possible to set other search paths for those document. The search-path is relative the package path.

NOTE: That the path is a relative filepath i.e both directory and file. Directory may be omitted.

For example, look for _package-overview.adoc_ in package folder instead of the default _overview.adoc_, _\_design/overview.adoc_:

```bash
goasciidoc --stdout -d package-overview.adoc 
```

## Thanks
The package `goparser` was taken from an open source project [by zpatrick](https://github.com/zpatrick/go-parser). It seemed abandoned so I've integrated it into this project (and extended it) and now it deviates rather much from it's earlier pure form ;). Many thanks @zpatrick!! That part has a [MIT License](https://github.com/zpatrick/go-parser/blob/master/LICENSE).

`copy.go` is created by Roland Singer [roland.singer@desertbit.com] and is used for unit test. Many thanks @r0l1. You may find the original [here](https://gist.github.com/r0l1/92462b38df26839a3ca324697c8cba04).

## Notes
This project consists of a parser to parse go-code and a producer to produce asciidoc files from the code & code documentation. It bases its rendering system heavily on templates (`asciidoc/template.go`) with some "sane" default so it may be rather easily overridden.

### List Default Templates

To list the default templates just do `goasciidoc --list-template`. Version 0.4.0 will list the following template names:

* interfaces
* interface
* consts
* typedeffunc
* package
* import
* typedefvars
* vars
* index
* function
* typedeffuncs
* functions
* structs
* struct
* typedefvars
* var
* const
* receivers

### Get Default Templates

It is possible to retrieve the default templates (_use list to get the template names_) using a command switch `--out-template NAME`, for example:

```bash
goasciidoc --out-template struct
``` 

The above outputs (for v0.0.6):

```
"=== {{.Struct.Name}}
[source, go]
----
{{.Struct.Decl}} {
{{- range .Struct.Fields}}
        {{if .Nested}}{{.Nested.Name}}{{"\t"}}struct{{else}}{{tabify .Decl}}{{end}}
{{- end}}
}
----

{{.Struct.Doc}}
{{range .Struct.Fields}}{{if not .Nested}}
==== {{.Decl}}
{{.Doc}}
{{- end}}
{{end}}
{{range .Struct.Fields}}{{if .Nested}}{{render $ .Nested}}{{end}}{{end}}
"
```

### Override Default Templates

If you're unhappy with one of the default templates, you may _override_ it (one or more) using the `-t FILEPATH` switch. It may be several `-t` on same command if multiple overrides. The filepath is either relative or fully qualified filepath to a template file.

For example, overriding the _package_ template can be done like this:
```bash
echo "== Override Package {{.File.FqPackage}}" > t.txt; goasciidoc -r package=t.txt --stdout; rm t.txt
```

In the `stdout` you may observe, now, it has _Override Package_ instead of _Package_ as heading

```
== Override Package github.com/mariotoffia/goasciidoc/goparser
=== Imports
...
```

### Override default templates using a files in a directory

It is possible to set a template directory where `goasciidoc` will search for files named (_see list templates_) and file extension _.gtpl_ e.g. _import.gtpl_.

Example usage: `goasciidoc --templatedir defaults`

It reads all files and overrides those found, the rest is using the default. You can checkout the _defaults_ folder (or copy as starting point) when you make your own layout. You can remove those not needed, and the defaults will kick in.

```bash
ls -l defaults
total 72
-rw-r--r-- 1 martoffi martoffi 104 Mar 19 21:33 const.gtpl
-rw-r--r-- 1 martoffi martoffi 256 Mar 19 21:33 consts.gtpl
-rw-r--r-- 1 martoffi martoffi 208 Mar 19 21:25 function.gtpl
-rw-r--r-- 1 martoffi martoffi 142 Mar 19 21:25 functions.gtpl
-rw-r--r-- 1 martoffi martoffi 159 Mar 19 21:33 import.gtpl
-rw-r--r-- 1 martoffi martoffi 623 Mar 19 21:24 index.gtpl
-rw-r--r-- 1 martoffi martoffi 307 Mar 19 21:26 interface.gtpl
-rw-r--r-- 1 martoffi martoffi 111 Mar 19 21:26 interfaces.gtpl
-rw-r--r-- 1 martoffi martoffi 220 Mar 19 21:24 package.gtpl
-rw-r--r-- 1 martoffi martoffi 148 Mar 19 21:27 receivers.gtpl
-rw-r--r-- 1 martoffi martoffi 562 Mar 19 21:26 struct.gtpl
-rw-r--r-- 1 martoffi martoffi 105 Mar 19 21:27 structs.gtpl
-rw-r--r-- 1 martoffi martoffi  92 Mar 19 21:32 typedeffunc.gtpl
-rw-r--r-- 1 martoffi martoffi 120 Mar 19 21:32 typedeffuncs.gtpl
-rw-r--r-- 1 martoffi martoffi 175 Mar 19 21:34 typedefvar.gtpl
-rw-r--r-- 1 martoffi martoffi 126 Mar 19 21:34 typedefvars.gtpl
-rw-r--r-- 1 martoffi martoffi 102 Mar 19 21:34 var.gtpl
-rw-r--r-- 1 martoffi martoffi 111 Mar 19 21:34 vars.gtpl
```

### Macros

There are a initial support for macros in _goasciidoc_, currently only _${gad:current:fq}_ is supported and will substitute the macro to the current fully qualified path to the source file. This can be e.g. used for inclusions of source code.

**Example Documentation**
```go
// ParseConfig to use when invoking ParseAny, ParseSingleFileWalker, and
// ParseSinglePackageWalker.
//
// .ParserConfig
// [source,go]
// ----
// include::${gad:current:fq}[tag=parse-config,indent=0]
// ----
// <1> These are usually excluded since many testcases is not documented anyhow
// <2> As of _go 1.16_ it is recommended to *only* use module based parsing
// tag::parse-config[]
type ParseConfig struct {
	// Test denotes if test files (ending with _test.go) should be included or not
	// (default not included)
	Test bool // <1>
	// Internal determines if internal folders are included or not (default not)
	Internal bool
	// UnderScore, when set to true it will include directories beginning with _
	UnderScore bool
	// Optional module to resolve fully qualified package paths
	Module *GoModule // <2>
}

// end::parse-config[]
```

It will then get rendered as follows:
![macro-expansion](https://raw.githubusercontent.com/mariotoffia/goasciidoc/master/docs/assets/macro-substitution.png)

### Plugins
Since asciidoc supports plugins, thus is **very** versatile, myself is using [kroki](https://kroki.io) that may render many types of diagrams (can be done online or offline using docker-compose). Below there are just a few of many, many [diagrams](https://kroki.io/examples.html) that may be outputted just using kroki.


For example a sequence diagram based on the following text in your documentation
```
sequenceDiagram
    participant Alice
    participant Bob
    Alice->John: Hello John, how are you?
    loop Healthcheck
        John->John: Fight against hypochondria
    end
    Note right of John: Rational thoughts prevail...
    John-->Alice: Great!
    John->Bob: How about you?
    Bob-->John: Jolly good!
```

will render the following sequence diagram
![sequence diagram sample](https://kroki.io/mermaid/svg/eNpljzFuwzAMRXefgtlrZzcKBymCtMjQoTegZdYiwoqqTLfw7SMrSDKEi4j_H6nPaqLfmYKjA-OY8KeCXBGTseOIwWAv7OhJfdO-aMWtu5P60MIHiSis_Qt4_QdMBIvOu0KKaswEinnnyZ2LuNbK3zYcefQGOCKHycAvUZ3XMCTGglMYyvupRpAKq99wHf1CYw0oYF7n7Ezw2qdtFxP9IUvTNNX9s7orsVt4T4S2eRhdPiufsUbvdbZH-KzXt4wnFVlgVB021QUIImOl)


or are you into packet diagrams will the following text
```
packetdiag {
  colwidth = 32;
  node_height = 72;

  0-15: Source Port;
  16-31: Destination Port;
  32-63: Sequence Number;
  64-95: Acknowledgment Number;
  96-99: Data Offset;
  100-105: Reserved;
  106: URG [rotate = 270];
  107: ACK [rotate = 270];
  108: PSH [rotate = 270];
  109: RST [rotate = 270];
  110: SYN [rotate = 270];
  111: FIN [rotate = 270];
  112-127: Window;
  128-143: Checksum;
  144-159: Urgent Pointer;
  160-191: (Options and Padding);
  192-223: data [colheight = 3];
}
```

render the packet diagram image
![packet diagram](https://kroki.io/packetdiag/svg/eNptkU9Pg0AQxe9-ijnqYRN2QSgYD6bGPzFpCW1jTGPMyk5hQ9mtsMjB-N0dIGk8cH2_mXkzb04yr9ApLQv4uQDI7bHXypVwC764IcFYhR8l6qJ0pEWkkegxfp3AxnZNjpDaxg2VPGQ-T-AeW6eNdNqaM_IFC31qwK8ODbWsuvoTm4GEAYtp1F1eGdsfURU1GvePxyGLYxoqnYT14dDiZOXRBh71Zdhi841qEsMEdtkj7BvrpENaV0Te-4Qi8li-zKJFAunmaRaRc7bZziHu0Tlvq1lEITw8zyPBuKBVXrVRth8lsWA8oGyWJeZV29WjGAQUMJnvmmKII7XauCkPHtLlMTlcrk9DxC1IoyCVSmlTXI0VsWBC0EQ1ZLanh56_59MWv39OCoi9)

Simple activity diagram can be annotated like this
```
actdiag {
  write -> convert -> image

  lane user {
    label = "User"
    write [label = "Writing reST"];
    image [label = "Get diagram IMAGE"];
  }
  lane actdiag {
    convert [label = "Convert reST to Image"];
  }
}
```

and outputs the following:
![Activity](https://kroki.io/actdiag/svg/eNrjquZSUCgvyixJVdC1U0jOzytLLSoBMTNzE9NTuYCSOYl5qQqlxalFCiClCiCBpNQcBVsFpVCgoBJEDGJCNFwqHMjPzEtXKEoNDlGKtYYoAhuJpMg9tUQhJTMxvSgxV8HT19HdFaKyFmZpYnIJSBpmL8xxCAOcoSIgWxRK8hU8QRbADKnlAgCssECm)

If you're into UML, you may use this annotation format
```
[Pirate|eyeCount: Int|raid();pillage()|
  [beard]--[parrot]
  [beard]-:>[foul mouth]
]

[<abstract>Marauder]<:--[Pirate]
[Pirate]- 0..7[mischief]
[jollyness]->[Pirate]
[jollyness]->[rum]
[jollyness]->[singing]
[Pirate]-> *[rum|tastiness: Int|swig()]
[Pirate]->[singing]
[singing]<->[rum]

[<start>st]->[<state>plunder]
[plunder]->[<choice>more loot]
[more loot]->[st]
[more loot] no ->[<end>e]

[<actor>Sailor] - [<usecase>shiver me;timbers]
```

to output this
![UML](https://kroki.io/nomnoml/svg/eNpdULFqwzAQ3fUVGpOCQ7eCY7R06hAodBQazvbFVpGlcHdqCfjja2FMnIKGp3fv3eOdsp-eQHDGO76nHKXWH1FmAt8fjuebDwEGPBxnpbVtEah3VWVvQJTE7bja2GvKQU8py-iUU8o20LIQdGIuQJB7JNfUi3nNc2oDlX49nd7s5LkbPV6XwXcK4R6R2VXmIX9iKU__KfZxWN5usdEvRTgLsPgiW7vxrx8Ox71u591Qs4UsRViAxLAUZfkImlvIsTRSdkNl1o3Jd2imRKhDKheyD1xinhkdky42jL3B9WSdJDJf4EMipyttm8zYAaPh0f8g6QnP4qcWiZ36A9RWnDM=)

You may also be a component fan
```
@startuml
!include C4_Container.puml

LAYOUT_TOP_DOWN
LAYOUT_WITH_LEGEND()

title Container diagram for Internet Banking System

Person(customer, Customer, "A customer of the bank, with personal bank accounts")

System_Boundary(c1, "Internet Banking") {
    Container(web_app, "Web Application", "Java, Spring MVC", "Delivers the static content and the Internet banking SPA")
    Container(spa, "Single-Page App", "JavaScript, Angular", "Provides all the Internet banking functionality to customers via their web browser")
    Container(mobile_app, "Mobile App", "C#, Xamarin", "Provides a limited subset of the Internet banking functionality to customers via their mobile device")
    ContainerDb(database, "Database", "SQL Database", "Stores user registration information, hashed auth credentials, access logs, etc.")
    Container(backend_api, "API Application", "Java, Docker Container", "Provides Internet banking functionality via API")
}

System_Ext(email_system, "E-Mail System", "The internal Microsoft Exchange system")
System_Ext(banking_system, "Mainframe Banking System", "Stores all of the core banking information about customers, accounts, transactions, etc.")

Rel(customer, web_app, "Uses", "HTTPS")
Rel(customer, spa, "Uses", "HTTPS")
Rel(customer, mobile_app, "Uses")

Rel_Neighbor(web_app, spa, "Delivers")
Rel(spa, backend_api, "Uses", "async, JSON/HTTPS")
Rel(mobile_app, backend_api, "Uses", "async, JSON/HTTPS")
Rel_Back_Neighbor(database, backend_api, "Reads from and writes to", "sync, JDBC")

Rel_Back(customer, email_system, "Sends e-mails to")
Rel_Back(email_system, backend_api, "Sends e-mails using", "sync, SMTP")
Rel_Neighbor(backend_api, banking_system, "Uses", "sync/async, XML/HTTPS")
@enduml
```

like so
![Component-Diagram](https://kroki.io/c4plantuml/svg/eNqVVdtu2zAMfc9XcNlLBrgtBuwDmhvWFkmb1enaPRm0zCRCZcmQ5LTFsH8f5dhOnBS7-MmiycPDQ1K-dB6tL3PV-yC1UGVGMP6SjI32KDXZ8yJ86s2GP-4elsnybpFM7h5vm_Pj9fIqmU2_Tm8ng0-9npdecXgTC5nEtcUcVsbCtfZkNXkYoX6Weg3xm_OU93oLss7ogSidNznZCMbtW38IjRnMCvyGIOXoCF6k30BRBaKqbIBCmFJ712ceO-hkxIYM7dtAfGasYwL9T_CzB_y0fAcvlCZYFOz8SCkMi0JJgV4a3WfTDW4xgriwgfz8-zjYJqTklmlU1FhILwUIhiPtAXVWmdu8aVP4Ysgku5ldwdj9mD8rOlvgmkL2JmssrCx8BEO9LhXaYF5Ys5UZOUCl3s-yKrUI1FFJ_wbesJKVkA62EkOItMD1QmrNiyN7wig3qVRUyzGvDg2n8ccInjBHVqLLBZTMpacMXJk65lL37F-4uRNyOwKQ0VYKOqY3SQcZekzRUWhD_RrYxN9m0Dl7Y5lbyTWCpbV03iKnBql5KvOquxFs2DkDLHmqhKWM2ydRuSgMFTkHyqz5QF6cn8iUongmnbFOMszr4vr9sZkYdrP7uI5uf9EnSMLAnPpXO9rTVz-gHKVKXGVguOnZnM_1WgX8JUsvK2jekbkU1jiz8jB9FRvUPGG7SIY9wKwZ7FEZU694h-lobw-kDSNYd1qwoa3iQGHA1JR-3-aoXdcIuB_aYVXvXuTePamDK2G_mA-OXMh9tVwuYnbs-u3W6M8-ncGuXHfpkluS601qDq6BHVyz5DVSZey2vUmI7k2LCG7iu9uLw-SHKf8rMhmx957YfuS7KPeEmYOVNXl167xYXkK-k0xAroEno3FTZ8A80ONojGKGdUBnwVph7IkcTVyXQzeudOGGbdPH8-WiBmqL6YSfzF0jTIi_qNV5ms9acS45lP9MvwGuTUmi)

It is even possible to generate a nice bar-chart like this (with some obscure JSON syntax ;)

![Vega-Lite](https://kroki.io/vegalite/svg/eNq9WE1v4zYQve-vGDAt3KKKrG_bC-QQIL0W3fYYGAtaomVuZMkRaW-MbP77jmhZlk3FklynOTgUOZx58x7FIfX6CYD8IsIFW1LyGchCypX4PBxuWEzNmMvFembybLgzUL23CZdsuPHMbyJLiVHMj5gIc76SHDvQxz1EfMPymKcxCEnDJxbBjOYQLmguYZ7lIFgq-RJ_BMjsO80jARQ7JWRzYDxeSHheM1H4EwY6E6uEbtEJFbBieYjzaMwEfEd4kLK1zGkCORMrNMdugc9RlBTR5YKB9Sssaf5UIqWySPMV2_i0oQmGwedH9QzwSvZxizy-lG2wiQFEbles6P1X5lkaJ9sCF41zxopB5QlHHQ8fDhixxzJHb0YP9w9NXiea24lp93L7F0OyWA7KN6SoQRN8Xw9kj02_V6T7U6f2xBmdevUn5qSX14p1DbM7CjTMdjvpTndNGyhxerl_-BCv3SS1tAXZK4impiZl4JrBZWRrWG27N1j3chmdfs4frkKu21_B_wpck1B7x70Lab4GVK-zgP3J9vq_hE77buf1l1Bb14Hp9grT_hp6rjm-jGsNrbafuqN2VvwP1NFv01Hj12vfS_3-MrqXLBf_fHHUyA6C9uLldxNSL-YdfAedhbxkVQe9pezr9MJNtYuUwVkpJ9ohpy_0M4echtNC6-s--sBXcnT90ji6wr7q9IvSXhrH1mUsX4GP8QfKN76-fOP_42w6PqvetTjusFtY1pvyPMVfFYTgrTMVeLldVvfJ1zI2CWkSrhMqVRQ-_w1voeulWYSFu7s7GGjiDoxbx7B-hz_gyBqNBw8HE7vBBB0O3id_YFjNc-4G97tx-73xY3IGRoGP7NkltLhFk-evWR6xnKjekvrOJFSJwY8f0IGh0uYgyzvI4SwfmhcYQlNqgscpi77WFsFRkkR95ygMaxZGOfmRbNRdduOQKf6L82y9mm3VQLUop6d0fct4SmMEGe8o23-iOJgoszlnSdQM0KjbZStltF4ed--Sy-ZzwSSpBt7K1rQioRnzMQVHAu9o3dhwWzJcxthzQtIXcnamc2amQwqECh1R33awd0bz3SceloZZxNP48JnnpWrWCUMAVXr7HeB5TVPJJZV8w2or4IWLmovCnstETfj7ZD3suStTIwgVJ9aDOvu0ybYRVsWuBi7Nlrgkkk64vpx6gUplPJcY9RnhU-FkThPBav0Y6s8XiakVJzCrNhBlS8rT_YzGpMMsyfLG5FQq7YklLGZp9E5q_5Qf-irCq2kC1xE7nlWhfWyuoPXSd75mVcXmpEpMa9xgBVDV4ZHchK4VOeq-dTN3qT9WF-ibUP2p5sQLg4iqpj0aWXR25GlPDe6nippjnoua8-ntJ7dCWRo=)