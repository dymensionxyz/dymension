// Packet uinv provides a simple way to define and register invariants in Cosmos SDK modules.
// Invariants should be written using normal code style, by returning errors containing context.
// Use NewErr to wrap errors in a way that they can be handled as invariant breaking.
package uinv
