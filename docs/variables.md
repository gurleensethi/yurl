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

Variables are sourced from a **variable set**. If a variable doesn't exist in the variable set, user is explicitly prompted to enter the value for that value.

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
        body: |
          {
            "id": {{ id }}
          }
    ```

    User will only be prompted for `id` once when `yurl` encounters the path `/todos/{{ id }}`, once the value is entered it is saved in the variable set and reused in body.
