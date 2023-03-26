# Home

![](media/yurl-icon.png){ width=300px  }

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

## Feature Rundown

Below is a quick (non-exhaustive) feature rundown of `yurl`.

### Runtime User Input

Use placeholders `{{ }}` to get input during runtime. For example `{{ id }}` will prompt to enter the id before executing the request.

```yaml title="http.yaml"
config: ...

requests:
  GetTodo:
    path: /todos/{{ id }}
    method: GET
```

```bash linenums="0"
yurl GetTodo
Enter `id`:
```

Your can use placeholders almost everywhere, in url path, request body, headers.

### Pre-Run Requests

You can pre run multiple requests before executing a request. Additionally, response from pre requests can be captured to be used in subsequent requests.

```yaml title="http.yaml"
config: ...

requests:
  Login:
    method: POST
    path: /auth/login
    jsonBody: |
      {
          "email": "{{ email }}"
          "password": "{{ password }}"
      }
    exports:
    authToken:
      json: $.authToken # (1)!

  GetUser:
    method: GET
    path: /me
    pre: # (2)!
      - name: Login
    headers:
      Authorization: Bearer {{ authToken }} # (3)!
```

1. Capture the `authToken` from response. Here we are using `jsonpath` to parse the json body response.
2. Defining a list of requests to pre-run before executing this request.
3. Using the "exported" `auth` from **Login** request. All the exported variables from a pre-request are automatically available for use.
