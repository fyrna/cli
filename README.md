## fyrna/cli

a cute litte cli framework

> [!WARNING]
> As long as main branch named "trunk", it's means breaking change can happen anytime.
>
> Don't use it now, wait for at least 0.1
>
> Also I don't accept any PR, just issue only

### Example

A simple "just works"
```go
package main

import (
    "fmt"
    
    "github.com/fyrna/cli"
)

func main() {
    app := cli.New()

    app.Command("cute", func(c *cli.Context) error {
        fmt.Println("hello cutie")
        return nil
    })

    app.Run()
}
```

Subcommand? i gotchu
```go
app.Command("cute miaw",
    cli.Short("say something good"),
    cli.Action(func(c *cli.Context) error {
        msg := "nothing"
        
        if len(c.Args()) > 0 {
            msg = c.Args().Get(1)
        }
        
        fmt.Prinltln("miaw says", msg)
        return nil
    }))
```
