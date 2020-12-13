# Event Bus in Go

Event Bus is a ***TCP*** server accepting multiple connections
that allows the clients to produce/consume events.

The events and streams are stored in ***BadgerDB***, which is embeddable
file database which makes the entire service absolutely portable.

The project uses classic MVC pattern supporting Go modules

### Operations

The event bus supports the following operations:

- `health`
- `create_stream`
- `delete_stream`
- `get_stream_info`
- `write_event`
- `process_events`
- `retry_events`
- `snapshot_db`
- `exit`

For better use of Event Bus make sure to use one of the client libs:

- [Event Bus Go Client](https://github.com/go-web-dev/event-bus-go-client)

### Samples

Here are some sample JSON requests for Event Bus.
So it's just a matter of copy and paste if you're using
something like `telnet` to talk over TCP.

```
{"operation": "health"}

{"operation": "create_stream", "body": {"stream_name": "steve"}}
{"operation": "create_stream", "body": {"stream_name": "john"}}
{"operation": "create_stream", "body": {"stream_name": "mike"}}

{"operation": "delete_stream", "body": {"stream_name": "steve"}}
{"operation": "delete_stream", "body": {"stream_name": "john"}}

{"operation": "get_stream_info", "body": {"stream_name": "steve"}}
{"operation": "get_stream_info", "body": {"stream_name": "john"}}

{"operation": "write_event", "body": {"stream_name": "steve", "event": {"e1": "event 1 value"}}}
{"operation": "write_event", "body": {"stream_name": "steve", "event": {"e2": "event 2 value"}}}
{"operation": "write_event", "body": {"stream_name": "steve", "event": {"e3": "event 3 value"}}}
{"operation": "write_event", "body": {"stream_name": "steve", "event": {"e4": "event 4 value"}}}
{"operation": "write_event", "body": {"stream_name": "steve", "event": {"e5": "event 5 value"}}}

{"operation": "write_event", "body": {"stream_name": "john", "event": {"e1": "event 1 value"}}}
{"operation": "write_event", "body": {"stream_name": "john", "event": {"e2": "event 2 value"}}}

{"operation": "process_events", "body": {"stream_name": "steve"}}
{"operation": "process_events", "body": {"stream_name": "john"}}

{"operation": "retry_events", "body": {"stream_name": "steve"}}

{"operation": "snapshot_db", "body": {"output": "backups"}}

{"operation": "exit"}
```
