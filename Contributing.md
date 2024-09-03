
## Code standards


### Errors

80% of production code, especially in Go, is error handling, so this is an important section.

**Guidelines**

1. Error strings should not interpolate data. Errors should have a CTRL+F'able prefix followed by colon seperated key value pairs of data.
2. Error strings should be all lower case.


1. Do not define new error types unless you need to handle a specific case.  Errors should wrap existing errors types in the [Dymension error library](https://pkg.go.dev/github.com/dymensionxyz/gerr-cosmos@v1.0.0/gerrc). This is based on the following principle from Google API guidelines

>(Google) APIs must use the canonical error codes defined by google.rpc.Code. Individual APIs must avoid defining additional error codes, since developers are very unlikely to write logic to handle a large number of error codes. For reference, handling an average of three error codes per API call would mean most application logic would just be for error handling, which would not be a good developer experience.

```

# WRONG
var 	ErrFulfillerAddressDoesNotExist = errorsmod.Register(ModuleName, 7, "Fulfiller address does not exist")
if (..){ return ErrFulfillerAddressDoesNotExist}

# RIGHT
if (..){ return errorsmod.Wrap(gerrc.ErrNotFound, "fulfiller")
```





Theref

2. Do not panic except in **trivial** cases. Instead of panicking return [gerrc.ErrInternal](https://pkg.go.dev/github.com/dymensionxyz/gerr-cosmos@v1.0.0/gerrc), wrapped with diagnostic info.

1. Error strings should not interpolate data. Errors should have a CTRL+F'able prefix followed by colon seperated key value pairs of data.


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