# pzlog

This is a pretty console writer for zerolog, powered by pterm.


## Usage

```go
  log := zerolog.New(pzlog.NewPtermWriter())
  log.Info().Msg("ok")
```

## Screenshot

![debug](./asset/debug.png)

![info](./asset/info.png)

![warn](./asset/warn.png)

![error](./asset/error.png)
