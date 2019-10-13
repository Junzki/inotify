# INotify

Go based Django-like signal utilities. 

[![codecov](https://codecov.io/gh/Junzki/inotify/branch/master/graph/badge.svg)](https://codecov.io/gh/Junzki/inotify)


## Usage
```golang
import (
	"fmt"

	"github.com/junzki/inotify"
)

func sigHandle(s inotify.ISignal, _ ...interface{}) {
	fmt.Printf("%s called.", s.Name())
}

func main() {
	sig := inotify.NewSignal("some-signal", sigHandle)

    sig.Send()  // Calls each handler function.
    
    // Console output: "some-signal called."
}
```
