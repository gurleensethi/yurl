# Variables

## Using Variables

To use variables use double curly brackets `{{ }}`.

```yaml title="http.yaml"
requests:
  GetTodo:
    path: /todos/{{ id }} # (1)!
    method: GET
```

1. `{{ id }}` will be replaced with whatever is contained in the variable `id`.

## User Input

Variables are sourced from a **variable set**. If a variable doesn't exist in the variable set, user is explicitly prompted to enter the value for that variable.

```yaml title="http.yaml"
requests:
  GetTodo:
    path: /todos/{{ id }}
    method: GET
```

In case the variable `id` doesn't exist in the variable set, user gets a prompt to enter the value.

```bash
$ yurl GetTodo
Enter `id`:
```

Variables can be used in `path`, `headers`, `body`.

```yaml title="http.yaml"
requests:
  UpdateTodo:
    path: /todos/{{ id }} # (1)!
    method: PUT
    headers:
      Authorization: Bearer {{ accessToken }} # (2)!
    json: | # (3)!
      {
        "title": "{{ title }}" 
      }
```

1. Using variables in path.
2. Using variable in headers.
3. Using variables in body.

This is perfect for the cases when each time you execute a request you want to provide values at runtime.

???+ info "Once the value of variable has been sourced from a user, it is added to the variable set."

    In the below example,

    ```yaml title="http.yaml"
    requests:
      GetTodo:
        path: /todos/{{ id }}
        method: PATCH
        jsonBody: |
          {
            "id": {{ id }}
          }
    ```

    User will only be prompted for `id` once when `yurl` encounters the path `/todos/{{ id }}`, once the value is entered it is saved in the variable set and reused in body.

## Command Line Variables

You can pass value for variables directly from command line using `-var` or `--variable` flag.

```yaml title="http.yaml"
requests:
  UpdateTodo:
    path: /todos/{{ id }}
    method: PUT
    jsonBody: |
      {
        "title": "{{ title }}"
      }
```

```bash linenums="0"
$ yurl -var id=10 UpdateTodo
```

You can pass as many variables as you want.

```bash linenums="0"
$ yurl -var id=10 -var title="Hello World" UpdateTodo
```

All the variables are added to the **variable set**.

## Variables from a file

You can define variables in a file and pass all of them at once using `-var-file` or `--variable-file`.

Variables in the file need to follow the pattern: `key=value`.

```text title="local.vars"
email=test@test.com
password=password
```

```yaml title="http.yaml"
requests:
  Login:
    method: POST
    path: /auth/login
    jsonBody: |
      {
        "email": "{{ email }}",
        "password": "{{ password }}"
      }
```

```bash linenums="0"
$ yurl -var-file local.vars Login
```

You can provide as many variable files as you want.

```bash linenums="0"
$ yurl -var-file local.vars -var-file staging.vars Login
```

All the variables are added to the **variable set**.

## Variable Types

You can define types on variables when the value is prompted from the user.

Use the pattern: `{{ <var name>:<type> }}` to enforce a type.

```yaml title="http.requests"
requests:
  GetTodoById:
    method: GET
    path: /todos/{{ id:int }} # (1)!
    headers:
      Accept: application/json
```

1. Types are defined using `<var name>:<type>` pattern.

```bash linenums="0"
$ yurl GetTodoById
Enter `id` (int): 10
```

In the above example, type `int` is defined for the variable `id`. When user is prompted for the value of `id` the required type is also displayed.

If the entered value is not valid, `yurl` exists immediately.

```bash linenums="0"
$ yurl GetTodoById
Enter `id` (int): not int

input for `id` must be of type int
```

### Supported Types

Currently the following types are supported:

- `string`
- `int`
- `float`
- `bool`
