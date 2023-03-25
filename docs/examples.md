# Examples

## Basic

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

## Post request with json body

```yaml title="http.yaml"
config:
  host: jsonplaceholder.typicode.com
  port: 443
  scheme: https

requests:
  CreateTodo:
    path: /todos
    method: POST
    jsonBody: |
      {
        "title": "{{ title }}"
      }
```

```bash title="bash"
yurl CreateTodo
```

## Specify a request file

Use the `-f` or `--file` to specify a requests file.

```yaml title="requests.yaml"
config:
  host: jsonplaceholder.typicode.com
  port: 443
  scheme: https

requests:
  CreateTodo:
    path: /todos
    method: POST
    jsonBody: |
      {
        "title": "{{ title }}"
      }
```

```bash title="bash"
yurl -f requests.yaml CreateTodo
```