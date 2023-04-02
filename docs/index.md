# Home

![](media/yurl-icon.png){ width=300px }

**Motivation:** I want to stay in the terminal so Postman is not an option. `curl` is quick and fast, but no way to save and customize requests. Here comes `yurl` :fontawesome-solid-bolt-lightning:.

Yurl allows you to define your http requests in `yaml` format and run them directly from command line.

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

```bash title="bash" linenums="0"
yurl GetTodo
```

![Basic Example](./media/example-basic.gif)
