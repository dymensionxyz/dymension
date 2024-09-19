
## Code standards

This doc applies to all Dymension golang repos.

Note: you should defer to the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) and [100 Go Mistakes](https://100go.co/#error-management) as next resources after this doc.

### Errors

80% of production code (especially in Go) is error handling, so doing errors well means 80% of the code is good.

**Guidelines**

1. Error strings should not interpolate data. Errors should have a CTRL+F'able prefix followed by colon seperated key value pairs of data.

```go
// bad
fmt.Errorf("the %s sat on the %s", "cat", "mat")
// good
fmt.Errorf("animal sat down: species: %s, surface: %s ", "cat", "mat")
```

2. Error strings should be all lower case.

```go
// bad
fmt.Errorf("The Cat sat on the Mat")
// good
fmt.Errorf("the cat sat on the mat")
```

3. Error strings should not have 'failed to' or 'could not' in them

```go
// bad
fmt.Errorf("failed to find foo")
// good
fmt.Errorf("find foo")
```

4. Error strings should be for humans and logs. Do not embed huge amounts of data in strings. If the error needs to be handled programatically it should be a struct and details can be obtained with `errors.As`. Never parse errors strings. Avoid converting errors to strings with .Error(). This includes when testing errors.



5. Errors should not be compared using string comparison. Use `errors.Is`. In the SDK use `errorsmod.IsOf`.

```go
// bad
strings.Contains(err.Error(), "foo")
// good
errors.Is(err, ErrFoo)
```

6. Avoid defining new errors if they are not required for special case handling using `errors.Is` or `errorsmod.IsOf`. Wrap an existing error instead using `fmt.Errorf`. In the SDK use `errorsmod.Wrapf`.

```go
// bad
ErrFoo = errors.New("foo")
ErrFooBar = errors.New("foo bar")
return ErrFooBar
// good
ErrFoo = errors.New("foo")
return errorsmod.Wrap(ErrFoo, "bar")
```

6. Use the Dymension cosmos error library as basis for errors. They are automatically converted to GRPC codes and registered. Choose the appropriate pre defined `gerrc.Err*`, considering API. This is based on the following principle from Google API guidelines

>(Google) APIs must use the canonical error codes defined by google.rpc.Code. Individual APIs must avoid defining additional error codes, since developers are very unlikely to write logic to handle a large number of error codes. For reference, handling an average of three error codes per API call would mean most application logic would just be for error handling, which would not be a good developer experience.

```go
// bad
return errors.New("not found")
// good
return gerrc.ErrNotFound
```

7. Do often wrap errors with calling context. Do not use the name of the calling function as context.

```go
// not great
func bar() error {
    err := foo()
    return err
}
// worse
func bar() error {
    err := foo()
    return errorsmod.Wrap(err, "bar")
}
// good
func bar() error {
    err := foo()
    return errorsmod.Wrap(err, "foo")
}
```

### Panicking

Do not write panics on chain unless the case is absolutely trivial. MustUnmarshal and similar in the SDK are examples of acceptable times to panic. Otherwise, return an error wrapping [gerrc.ErrInternal](https://pkg.go.dev/github.com/dymensionxyz/gerr-cosmos@v1.0.0/gerrc#pkg-variables).



### SDK DoS Prevention

**Guidelines**


### SDK DoS Prevention

**Guidelines**

-

### References

- https://github.com/dymensionxyz/sdk-utils : Dymension golang and SDK utils library
- https://github.com/dymensionxyz/gerr-cosmos : Dymension error library
- https://cloud.google.com/apis/design/errors#error_codes : Google error handling guidelines
- https://github.com/uber-go/guide/blob/master/style.md#errors : Uber Style Guide