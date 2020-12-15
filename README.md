# Event Bus in Go

Event Bus is a ***TCP*** server accepting multiple connections
that allows the clients to produce/consume events.

The events and streams are stored in [***BadgerDB***](https://dgraph.io/docs/badger/get-started/), which is an embeddable
file database which makes the entire service absolutely portable.

The project uses classic MVC pattern and supports Go modules for dependency versioning as well.

**Note**:
At the moment of writing this micro service ***Go 1.15*** was used.

### Build

```
# install all required dependencies
go get ./...
```

```sh
# compile, build and run the program
go run main.go
```

```sh
# compile and build the program
go build

# run the program
./event-bus
```

### Test

```sh
# run unit tests
go test ./...

# run tests with coverage
./coverage.sh
```

### Operations

The event bus supports the following operations:

- `health`
- `create_stream`
- `delete_stream`
- `get_stream_info`
- `get_stream_events`
- `write_event`
- `process_events`
- `retry_events`
- `mark_event`
- `exit`

For better use of Event Bus make sure to use one of the client libs:

- [Event Bus Go Client](https://github.com/go-web-dev/event-bus-go-client)

**Note:**

Before making any kind of requests have a look inside `config/config.yaml`
in the `auth` section. Make sure to adjust `client_id` and `client_secret` or
simply add a new entry for another client.

### Sample Requests

Here are some sample JSON requests for Event Bus.
So it's just a matter of copy and paste if you're using
something like `telnet` to talk over TCP.

```
{"operation": "health"}

{"operation": "create_stream", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "steve"}}
{"operation": "create_stream", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "john"}}
{"operation": "create_stream", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "mike"}}

{"operation": "delete_stream", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "steve"}}
{"operation": "delete_stream", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "john"}}

{"operation": "get_stream_info", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "steve"}}
{"operation": "get_stream_info", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "john"}}

{"operation": "get_stream_events", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "steve"}}
{"operation": "get_stream_events", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "john"}}

{"operation": "write_event", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "steve", "event": {"e1": "event 1 value"}}}
{"operation": "write_event", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "steve", "event": {"e2": "event 2 value"}}}
{"operation": "write_event", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "steve", "event": {"e3": "event 3 value"}}}
{"operation": "write_event", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "steve", "event": {"e4": "event 4 value"}}}
{"operation": "write_event", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "steve", "event": {"e5": "event 5 value"}}}

{"operation": "write_event", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "john", "event": {"e1": "event 1 value"}}}
{"operation": "write_event", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "john", "event": {"e2": "event 2 value"}}}

{"operation": "process_events", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "steve"}}
{"operation": "process_events", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "john"}}

{"operation": "retry_events", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "steve"}}
{"operation": "retry_events", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"stream_name": "john"}}

{"operation": "mark_event", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"event_id": "1b78ca67-d916-4059-a299-1ebf49664eb2", "status": 1}}
{"operation": "mark_event", "auth": {"client_id": "dope_go_client_id", "client_secret": "dope_go_client_secret"}, "body": {"event_id": "7f85f2c4-4a5d-42fb-b24e-30ec894e053a", "status": 2}}

{"operation": "exit"}
```
