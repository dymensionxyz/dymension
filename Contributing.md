
## Code standards


### Errors

80% of production code, especially in Go, is error handling, so this is an important section.

**Guidelines**

- Do not define new error types unless you need to handle a specific case.  Errors should 
- Do not panic except in **trivial** cases. Instead of panicking 


### SDK DoS Prevention

**Guidelines**

- 

### References

- https://github.com/dymensionxyz/sdk-utils : Dymension golang and SDK utils library
- https://github.com/dymensionxyz/gerr-cosmos : Dymension error library
- https://cloud.google.com/apis/design/errors#error_codes : Google error handling guidelines
- https://github.com/uber-go/guide/blob/master/style.md#errors : Uber Style Guide