# Concurrent Clients

In this stage, you'll add support for **multiple concurrent clients**.

In addition to handling multiple commands from the same client, **Redis servers** are also designed to handle multiple clients at once.

To implement this, you'll need to either use **threads**, or, if you're feeling adventurous, an **Event Loop** (like the official Redis implementation does).

## Tests

The tester will execute your program like this:

```bash
$ ./your_program.sh
```

It'll then send two `PING` commands concurrently using two different connections:

```bash
# These two will be sent concurrently so that we test your server's ability to handle concurrent clients.
$ redis-cli PING
$ redis-cli PING
```

The tester will expect to receive two `+PONG\r\n` responses.

## Notes

Since the tester client only sends the `PING` command at the moment, it's okay to ignore what the client sends and hardcode a response. We'll get to parsing client input in later stages.