# Examples

## Basic

**Use Case**: Making a simple login POST request.

By default `yurl` loads in a file named `http.yaml`.

```yaml title="http.yaml"
config:
  host: jsonplaceholder.typicode.com
  port: 443
  scheme: https

requests:
  GetTodo:
    path: /todos/{{ id }}
    method: GET
```

```bash title="bash"
yurl GetTodo
```

![Basic Example](./media/example-basic.gif)

### Verbose output

```bash title="bash"
yurl -v GetTodo
```

![Basic Example](./media/example-basic-verbose.gif)
