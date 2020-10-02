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

Maybe you want to have a nice barchart, provide with the following json
```json
{
  "$schema": "https://vega.github.io/schema/vega-lite/v4.json",
  "description": "A diverging stacked bar chart for sentiments towards a set of eight questions, displayed as percentages with neutral responses straddling the 0% mark",
  "data": {
    "values": [
      {"question": "Question 1", "type": "Strongly disagree", "value": 24, "percentage": 0.7},
      {"question": "Question 1", "type": "Disagree", "value": 294, "percentage": 9.1},
      {"question": "Question 1", "type": "Neither agree nor disagree", "value": 594, "percentage": 18.5},
      {"question": "Question 1", "type": "Agree", "value": 1927, "percentage": 59.9},
      {"question": "Question 1", "type": "Strongly agree", "value": 376, "percentage": 11.7},
      {"question": "Question 2", "type": "Strongly disagree", "value": 2, "percentage": 18.2},
      {"question": "Question 2", "type": "Disagree", "value": 2, "percentage": 18.2},
      {"question": "Question 2", "type": "Neither agree nor disagree", "value": 0, "percentage": 0},
      {"question": "Question 2", "type": "Agree", "value": 7, "percentage": 63.6},
      {"question": "Question 2", "type": "Strongly agree", "value": 11, "percentage": 0},
      {"question": "Question 3", "type": "Strongly disagree", "value": 2, "percentage": 20},
      {"question": "Question 3", "type": "Disagree", "value": 0, "percentage": 0},
      {"question": "Question 3", "type": "Neither agree nor disagree", "value": 2, "percentage": 20},
      {"question": "Question 3", "type": "Agree", "value": 4, "percentage": 40},
      {"question": "Question 3", "type": "Strongly agree", "value": 2, "percentage": 20},
      {"question": "Question 4", "type": "Strongly disagree", "value": 0, "percentage": 0},
      {"question": "Question 4", "type": "Disagree", "value": 2, "percentage": 12.5},
      {"question": "Question 4", "type": "Neither agree nor disagree", "value": 1, "percentage": 6.3},
      {"question": "Question 4", "type": "Agree", "value": 7, "percentage": 43.8},
      {"question": "Question 4", "type": "Strongly agree", "value": 6, "percentage": 37.5},
      {"question": "Question 5", "type": "Strongly disagree", "value": 0, "percentage": 0},
      {"question": "Question 5", "type": "Disagree", "value": 1, "percentage": 4.2},
      {"question": "Question 5", "type": "Neither agree nor disagree", "value": 3, "percentage": 12.5},
      {"question": "Question 5", "type": "Agree", "value": 16, "percentage": 66.7},
      {"question": "Question 5", "type": "Strongly agree", "value": 4, "percentage": 16.7},
      {"question": "Question 6", "type": "Strongly disagree", "value": 1, "percentage": 6.3},
      {"question": "Question 6", "type": "Disagree", "value": 1, "percentage": 6.3},
      {"question": "Question 6", "type": "Neither agree nor disagree", "value": 2, "percentage": 12.5},
      {"question": "Question 6", "type": "Agree", "value": 9, "percentage": 56.3},
      {"question": "Question 6", "type": "Strongly agree", "value": 3, "percentage": 18.8},
      {"question": "Question 7", "type": "Strongly disagree", "value": 0, "percentage": 0},
      {"question": "Question 7", "type": "Disagree", "value": 0, "percentage": 0},
      {"question": "Question 7", "type": "Neither agree nor disagree", "value": 1, "percentage": 20},
      {"question": "Question 7", "type": "Agree", "value": 4, "percentage": 80},
      {"question": "Question 7", "type": "Strongly agree", "value": 0, "percentage": 0},
      {"question": "Question 8", "type": "Strongly disagree", "value": 0, "percentage": 0},
      {"question": "Question 8", "type": "Disagree", "value": 0, "percentage": 0},
      {"question": "Question 8", "type": "Neither agree nor disagree", "value": 0, "percentage": 0},
      {"question": "Question 8", "type": "Agree", "value": 0, "percentage": 0},
      {"question": "Question 8", "type": "Strongly agree", "value": 2, "percentage": 100}
    ]
  },
  "transform": [
    {
      "calculate": "if(datum.type === 'Strongly disagree',-2,0) + if(datum.type==='Disagree',-1,0) + if(datum.type =='Neither agree nor disagree',0,0) + if(datum.type ==='Agree',1,0) + if(datum.type ==='Strongly agree',2,0)",
      "as": "q_order"
    },
    {
      "calculate": "if(datum.type === 'Disagree' || datum.type === 'Strongly disagree', datum.percentage,0) + if(datum.type === 'Neither agree nor disagree', datum.percentage / 2,0)",
      "as": "signed_percentage"
    },
    {"stack": "percentage", "as": ["v1", "v2"], "groupby": ["question"]},
    {
      "joinaggregate": [
        {
          "field": "signed_percentage",
          "op": "sum",
          "as": "offset"
        }
      ],
      "groupby": ["question"]
    },
    {"calculate": "datum.v1 - datum.offset", "as": "nx"},
    {"calculate": "datum.v2 - datum.offset", "as": "nx2"}
  ],
  "mark": "bar",
  "encoding": {
    "x": {
      "field": "nx",
      "type": "quantitative",
      "axis": {
        "title": "Percentage"
      }
    },
    "x2": {"field": "nx2"},
    "y": {
      "field": "question",
      "type": "nominal",
      "axis": {
        "title": "Question",
        "offset": 5,
        "ticks": false,
        "minExtent": 60,
        "domain": false
      }
    },
    "color": {
      "field": "type",
      "type": "nominal",
      "legend": {
        "title": "Response"
      },
      "scale": {
        "domain": ["Strongly disagree", "Disagree", "Neither agree nor disagree", "Agree", "Strongly agree"],
        "range": ["#c30d24", "#f3a583", "#cccccc", "#94c6da", "#1770ab"],
        "type": "ordinal"
      }
    }
  }
}
```

and it will output a nice barchart like this
![Vega-Lite](https://kroki.io/vegalite/svg/eNq9WE1v4zYQve-vGDAt3KKKrG_bC-QQIL0W3fYYGAtaomVuZMkRaW-MbP77jmhZlk3FklynOTgUOZx58x7FIfX6CYD8IsIFW1LyGchCypX4PBxuWEzNmMvFembybLgzUL23CZdsuPHMbyJLiVHMj5gIc76SHDvQxz1EfMPymKcxCEnDJxbBjOYQLmguYZ7lIFgq-RJ_BMjsO80jARQ7JWRzYDxeSHheM1H4EwY6E6uEbtEJFbBieYjzaMwEfEd4kLK1zGkCORMrNMdugc9RlBTR5YKB9Sssaf5UIqWySPMV2_i0oQmGwedH9QzwSvZxizy-lG2wiQFEbles6P1X5lkaJ9sCF41zxopB5QlHHQ8fDhixxzJHb0YP9w9NXiea24lp93L7F0OyWA7KN6SoQRN8Xw9kj02_V6T7U6f2xBmdevUn5qSX14p1DbM7CjTMdjvpTndNGyhxerl_-BCv3SS1tAXZK4impiZl4JrBZWRrWG27N1j3chmdfs4frkKu21_B_wpck1B7x70Lab4GVK-zgP3J9vq_hE77buf1l1Bb14Hp9grT_hp6rjm-jGsNrbafuqN2VvwP1NFv01Hj12vfS_3-MrqXLBf_fHHUyA6C9uLldxNSL-YdfAedhbxkVQe9pezr9MJNtYuUwVkpJ9ohpy_0M4echtNC6-s--sBXcnT90ji6wr7q9IvSXhrH1mUsX4GP8QfKN76-fOP_42w6PqvetTjusFtY1pvyPMVfFYTgrTMVeLldVvfJ1zI2CWkSrhMqVRQ-_w1voeulWYSFu7s7GGjiDoxbx7B-hz_gyBqNBw8HE7vBBB0O3id_YFjNc-4G97tx-73xY3IGRoGP7NkltLhFk-evWR6xnKjekvrOJFSJwY8f0IGh0uYgyzvI4SwfmhcYQlNqgscpi77WFsFRkkR95ygMaxZGOfmRbNRdduOQKf6L82y9mm3VQLUop6d0fct4SmMEGe8o23-iOJgoszlnSdQM0KjbZStltF4ed--Sy-ZzwSSpBt7K1rQioRnzMQVHAu9o3dhwWzJcxthzQtIXcnamc2amQwqECh1R33awd0bz3SceloZZxNP48JnnpWrWCUMAVXr7HeB5TVPJJZV8w2or4IWLmovCnstETfj7ZD3suStTIwgVJ9aDOvu0ybYRVsWuBi7Nlrgkkk64vpx6gUplPJcY9RnhU-FkThPBav0Y6s8XiakVJzCrNhBlS8rT_YzGpMMsyfLG5FQq7YklLGZp9E5q_5Qf-irCq2kC1xE7nlWhfWyuoPXSd75mVcXmpEpMa9xgBVDV4ZHchK4VOeq-dTN3qT9WF-ibUP2p5sQLg4iqpj0aWXR25GlPDe6nippjnoua8-ntJ7dCWRo=)