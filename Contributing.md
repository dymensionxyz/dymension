
## Code standards

This doc applies to all Dymension golang repos.

Note: you should defer to the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) and [100 Go Mistakes](https://100go.co/#error-management) as next resources after this doc.

### Errors

80% of production code (especially in Go) is error handling, so doing errors well means 80% of the code is good, which is a huge win.

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

4. Error strings should be for humans and logs. Do not embed large amounts of data in strings. If the error needs to be handled programatically it should be a struct and details can be obtained with `errors.As`. Never parse errors strings. Avoid converting errors to strings with .Error(). This includes when testing errors.



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

7. Use the Dymension cosmos error library as basis for errors. They are automatically converted to GRPC codes and registered. Choose the appropriate pre defined `gerrc.Err*`, considering API. This is based on the following principle from Google API guidelines

>(Google) APIs must use the canonical error codes defined by google.rpc.Code. Individual APIs must avoid defining additional error codes, since developers are very unlikely to write logic to handle a large number of error codes. For reference, handling an average of three error codes per API call would mean most application logic would just be for error handling, which would not be a good developer experience.

```go
// bad
return errors.New("not found")
// good
return gerrc.ErrNotFound
```

8. Do often wrap errors with calling context. Do not use the name of the calling function as context.

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

9. Do not log and return an error. Do one or the other

```go
// bad
func bar() error {
    err := foo()
    if err!=nil{
        log.Error("something", err)
    }
    return err
}
// good - case 1: return
func bar() error {
    err := foo()
    return err
}
// good - case 2: log
func bar() error {
    err := foo()
    if err!=nil{
        log.Error("something", err)
    }
    return nil
}
```

Note: it may be acceptable to do both at **outer** API boundaries (e.g. just before returning http response), if it helps debugging.

### Panics

Do not write panics on chain unless the case is absolutely trivial. MustUnmarshal and similar in the SDK are examples of acceptable times to panic. In case of invariant breakage/logic bug return an error wrapping [gerrc.ErrInternal](https://pkg.go.dev/github.com/dymensionxyz/gerr-cosmos@v1.0.0/gerrc#pkg-variables).

### Channels

Avoid go channels which do not have size 0 or 1. A valid use case for another size is if the consumer is a worker pool with fixed resources and you are OK blocking the producer when the limit is reached.

### Ctx

1. Avoid putting ctx objects in structs. They should be passed as first argument to functions only.


2. If putting a key in a ctx, use a private struct instance as the key (not a string).

### Micro-optimization

Don't (micro) optimize code unless it's a proven bottleneck. Favour terseness and readabilty.

```go
// bad
if len(foos) == 0 {
    return
}
for _, f := range foo {
    // ..
}
// good
for _, f := range foo {
    // ..
}
```

### Interfaces

1. Interfaces should always be defined in the package of the API consumer. Ubiquitious interfaces likes std `Stringer` are exceptions.


2. Interfaces should be named with verbs, not with `*I` suffix.

### Package design

1. Make small packages for specific things and abuse the naming convention in golang to your benefit.

```go
// bad
.. = utils.WidgetToGadget(..)
// better
.. = widget.ToGadget(..)
```

2. Make use of `/internal` subdirectories in packages to prevent things being exported and to shrink namespaces

### Utilies and shared code

It's encouraged to add golang and cosmos utilities and shared code to [dymensionxyz/sdk-utils](https://github.com/dymensionxyz/sdk-utils). Such code should be genuinely general purpose, well tested and documented.

### Proto and API definition

1. All protobuf fields should have docstrings. The docstrings don't need to follow a particular format, but they DO need to explain the field in the context of the business logic.

```protobuf
// ~~~~~
// bad (it's plainly obvious and useless)

// Amount of tokens claimed by user.
int claimed_tokens = 1;

// ~~~~~
// good

// Amount of tokens of original denom given to user as calculated by his shares of the alternate denom
int claimed_tokens = 1;

```

2. Do not reuse protobuf field numbers when updating protos. When in doubt, reserve the old number and use a new one. This applies even if the field is new! It just makes things a lot easier to review.

### Comments and documentation 

#### General

1. Do not write obvious comments or repeat the code in English.
2. Explain WHY, not how.
3. Do not comment out code. Code is in git. You can comment a github link to a region of code at a particular commit if needed.\
There are several reasons for this: commented out code doesn't survive refactors well, and clutters up searches. It wastes time.

#### Docstrings

1. Do not describe algorithm/logic internals in docstrings. Docstrings should describe contracts and APIs.
2. Do not worry about making the first word of the docstring on a public object the same as the object. Feel free to disable any linters/highlighters that pick up on it.

In other words, we are fine with the first version here:


```go
// bla bla
type Foo struct{ }

// Foo bla bla
type Foo struct{ }
```



### References

- https://github.com/dymensionxyz/sdk-utils : Dymension golang and SDK utils library
- https://github.com/dymensionxyz/gerr-cosmos : Dymension error library
- https://cloud.google.com/apis/design/errors#error_codes : Google error handling guidelines
- https://github.com/uber-go/guide/blob/master/style.md#errors : Uber Style Guide
- https://100go.co/#error-management : 100 Go Mistakes

### Appendix - instructions for AI tools

These are instructions for AI coding tools: LLM model and ML based tools. We want you to play the role of an expert software developer, who has particular expertise in Golang and its nuances and quirks, as well as the Cosmos SDK, IBC protocol, Tendermint/CometBFT algorithms and light clients, and also knowledge of the wider blockchain technology space. Please be terse, and follow all the coding guidelines in this document. Pay particular attention to security, and to make sure that the code is deterministic. Make sure to look at the context of the code to use the right library versions, and to follow typical Cosmos SDK practices like using the SDK math, coin and error handling libraries.