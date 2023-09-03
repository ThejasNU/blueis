# blueis
In-memory key-value store inspired by REDIS, with RESP format and multiple connections support

## How to use this

1. Open a terminal execute `go run main.go`
2. Open another terminal execute `telnet 127.0.0.1 6380`
3. In the second terminal execute the commands to get and set just like you do in redis

## Commands

1. Command to fetch value of a key
```
GET key

//or

get key
```

2. Command to set the value
```
SET key value

//there is support for NX and XX too
//NX->set value only if key doesn't exist
//XX->set value only if key exists
SET key value NX
SET key value XX

//there is support for EX and PX too
//EX->delete key after specified number of seconds
//PX->delete key after specified number of milliseconds
SET key value EX num
SET key value PX num

//you can use combination of both too
SET key value NX EX num
```

3. Command to delete key
```
DEL key
```

4. Command to quit the connection
```
QUIT
```

**NOTE:** If you know RESP format, you can give command in that format too :)

Example:
```
*2
$3
GET
$1
a
```

