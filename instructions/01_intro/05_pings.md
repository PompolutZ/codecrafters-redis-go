# Multiple PING Commands

In this stage, you'll respond to **multiple PING commands** sent by the same connection.

A **Redis server** starts to listen for the next command as soon as it's done responding to the previous one. This allows Redis clients to send multiple commands using the same connection.

## Tests

The tester will execute your program like this:

```bash
$ ./your_program.sh
```

It'll then send multiple `PING` commands using the same connection. For example, it might send:

```bash
$ echo -e "PING\nPING" | redis-cli
```

The tester will expect to receive multiple `+PONG\r\n` responses (one for each command sent).

You'll need to run a **loop** that reads input from a connection and sends a response back.

## Notes

- Just like the previous stage, you can hardcode `+PONG\r\n` as the response for this stage. We'll get to parsing client input in later stages.
- The `PING` commands will be sent using the **same connection**. We'll get to handling multiple connections in later stages.