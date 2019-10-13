# INotify

Go based Django-like signal utilites. 

## Usage
```golang
import (
    "fmt"
    "github.com/junzki/inotify"
)


func handlerFunc(s ISigal, _ ...interface{}) {
    fmt.Printf("%s called.\n", s.Name())
}

NamedSignal := inotify.NewSignal("NamedSignal", handlerFunc)


NamedSignal.Send()  // Calls each handler function.
// Console output: NamedSignal calld.
```
