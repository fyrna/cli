# fyrna/cli

✨ A cute little CLI framework written in Go 💕
Designed to be tiny, flexible, and cute~

> [!WARNING]
> As long as the main branch is named `trunk`, **breaking changes** may happen at any time!
>
> Please **don't use it yet!** Wait until version `v0.1.0` or later.
>
> Also... I don't accept PRs (yet~), but feel free to open issues! ✨

---

## 💻 Example Usage

### Basic command~!

```go
package main

import (
    "fmt"
    "github.com/fyrna/cli"
)

func main() {
    app := cli.New()

    app.Command("cute", func(ctx *cli.Context) error {
        fmt.Println("hello cutie")
        return nil
    })

    app.Run()
}
```

### Subcommand? Mew got you~

```go
    app.Command("cute miaw", func(ctx *cli.Context) error {
        msg := "nothing"

        if len(ctx.Args()) > 0 {
            msg = ctx.Args().Get(0)
        }

        fmt.Println("miaw says", msg)
        return nil
    },
    cli.Short("say something good"),
    cli.Usage("miaw <string>"))
```


---

## 📦 Box

- No dependency (zero deps, just Go stdlib!) 🍃
- Minimal and cute syntax (≧▽≦)
- Easy to use, but flexible for devs who love control 🎮
- Inspired by popular tools, but `fyrna/cli` cuter~ (uwu)

---

## 🌸 License

MIT~ do whatever you want as long as you’re being kind >w<
