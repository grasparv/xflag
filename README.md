xflag is a tiny command line parser

For the user:
* Options are prefixed with `-` and are _optional_
  * Options are given _after_ the arguments since this makes it easier to edit command lines
* Arguments are given positionally and are _mandatory_
* A brief help page is rendered automatically
  * Arguments and options have a clear type (string, bool, duration, etc)
* Avoids making the command-line parsing a complex programming language
  * The help text is a simple, single line of text
* Arguments and options are validated
  * All arguments/options must have the correct type
  * No mandatory arguments may be missing

Usage example:

```
$ myprog

       eval.......Build evaluation database from source files
    inspect.......Inspect database
      fetch.......Fetch remote data from upstream
    analyse.......Analyse sensor data in database

$ myprog eval -help

usage: eval <dumpfile> <dbfile> [-debug]

    <dumpfile>.......Path to eval dump    string
      <dbfile>.......Path to database     string
      [-debug].......Turn on debugging    bool    (default: "false")
```

As a programmer:
* It produces readable code and is generally non-intrusive.
* Construct options by make fields pointers, otherwise they are interpreted as arguments.
  * Defaults can be provided before the first comma.
* Values are supplied directly into the struct you define.

Code example:

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
	XFlag string `xflag:"eval|Evaluate the situation"`

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
