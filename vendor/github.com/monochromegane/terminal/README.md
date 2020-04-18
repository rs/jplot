# terminal

This is Go package that provides function like a `isatty()`.

## Installation

```sh
$ go get github.com/monochromegane/terminal
```

## Usage

```go
import (
        "github.com/monochromegane/terminal"
        "os"
)
if terminal.IsTerminal(os.Stdout) {
        // Do something.
}
```

## Support OS

- Mac OS X (CGO and Not CGO)
- Linux
- Windows
- FreeBSD (CGO and Not CGO)

## Contributing

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
4. Push to the branch (git push origin my-new-feature)
5. Create new Pull Request

