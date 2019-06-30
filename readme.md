go-xeger
=====

go-xeger is a golang module that generates random strings from a regular expression.

### Installation

To install go-xeger:

```bash
$ go get github.com/takahiromiyamoto/go-xeger
```

### Usage

```go
package main

import (
  "fmt"

  "github.com/takahiromiyamoto/go-xeger"
)

func main() {
  x, err := xeger.NewXeger("[0-9]+")
  if err != nil {
    panic(err)
  }

  fmt.Println(x.Generate())
}
```

### Contributing

Contributions are very welcome. Please open a tracking issue or pull request and we can work to get things merged in.
