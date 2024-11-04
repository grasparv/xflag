xflag is a tiny command line parser

* It is not advanced like cobra, but readable calling code and non-intrusive
* Fields with pointers are optional, otherwise they are mandatory arguments
  * Defaults can be provided before the first comma

```
package main

import (
	"fmt"
	"os"
	"time"

    "github.com/grasparv/xflag/v2"
)

type EvalCmd struct {
	// sub command and its description
	XFlag string `xflag:"evaldb|Evaluate the situation"`

	// arguments and flags
	Debug   *bool          `xflag:"false|Turn on debugging"`
	EvalDB  string         `xflag:"Path to eval database"`
	Timeout *time.Duration `xflag:"10s|Specify how long to run at most"`
}

type InspectCmd struct {
	// sub command and its description
	XFlag string `xflag:"inspect|Inspect the database"`

	// arguments and flags
	Debug    *bool  `xflag:"false|Turn on debugging"`
	Database string `xflag:"Path to the database"`
}

func main() {
	commands := []interface{}{
		EvalCmd{},
		InspectCmd{},
	}

	cmd, err := xflag.Parse(commands, os.Args)
	if err != nil {
		fmt.Print(err)
		return
	}

	switch cmd := cmd.(type) {
	case *EvalCmd:
		// do stuff
	case *InspectCmd:
		// do other stuff
	default:
		fmt.Print(xflag.GetUsage(commands))
	}
}
```
