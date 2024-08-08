# Notes-1

## recover()

```golang
func ListenForWs(conn *WebSocketConnection) {
 defer func() {
  if r := recover(); r != nil {
   log.Println("Error - ", fmt.Sprintf("%+v", r))
  }
 }()
}
```

- `recover()` function is provided by Go's standard library. It is a built-in function that is used in conjunction with defer to handle panics in a controlled manner.

- In Go, a panic can be triggered by the panic() function or by certain runtime errors (such as attempting to access an out-of-bounds slice index). When a panic occurs, the program stops executing normal code and starts unwinding the stack, running any deferred functions along the way.

- If `recover()` is called within a deferred function, it can catch the panic, allowing the program to continue running or handle the error gracefully. However, if `recover()` is not used, the panic will cause the program to terminate.

## deleting keys from map

- `delete(clients, client)`: we can use this syntax to delete a key from a map
