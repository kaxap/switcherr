# switcherr
Work in progress.

A linter for switch-case error handling.

Typical use case:

```go
a, err := someFunc()
switch {
case a == 0:
  return ErrNotFound
case err != nil:
  return err
}
```

In the code above `err != nil` is never reached due to default value of `a` being `0`.
