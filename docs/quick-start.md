# Quick Start

This document will walk you through the process of using yurl and explore its various features.

We will use the [JSONPlaceholder](https://jsonplaceholder.typicode.com/) API as an example, which is publically available, so you can follow along with the examples on your own.

## Basic

Let's start with the most basic example.

Create a `http.yaml` file with the following content:

```yaml title="http.yaml"
config:
  host: jsonplaceholder.typicode.com
  port: 443
  scheme: https

requests:
  ListTodos:
    method: GET
    path: /todos
```

We define a single request named `ListTodos` which will list all the todos from the [JSONPlaceholder](https://jsonplaceholder.typicode.com/) API.

Now, let's run the following command:

```bash
yurl ListTodos
```

`yurl` by default detects the `http.yaml` file and executes the `ListTodos` request.

![](./media/quick-start/quick-start-basic.gif)

## Request with user input

Now we define a request to fetch a single todo by its id. We will define the `id` parameter as a variable.

```yaml title="http.yaml" hl_lines="11-13"
config:
  host: jsonplaceholder.typicode.com
  port: 443
  scheme: https

requests:
  ListTodos:
    method: GET
    path: /todos

  GetTodo:
    method: GET
    path: /todos/{{ id }}
```

We define a single request named `GetTodo` which will fetch a single todo by its id. We use variables to get the `id` from the user.

![](./media/quick-start/quick-start-user-input.gif)

Learn more about variables [here](./variables.md).

## Verbose Mode

You can run the requests in verbose mode by passing the `-v` flag.

```bash
yurl -v GetTodo
```

![](./media/quick-start/quick-start-verbose-mode.gif)

## List Requests

We can list all the requests defined in the `http.yaml` file by running `yurl ls` command.

```bash
yurl ls
```

![](./media/quick-start/quick-start-list-requests.gif)

## Post Request with JSON Body

Let's define a request to create a new todo. We will use the `POST` method and pass a JSON body as the payload.

```yaml title="http.yaml" hl_lines="15-21"
config:
  host: jsonplaceholder.typicode.com
  port: 443
  scheme: https

requests:
  ListTodos:
    method: GET
    path: /todos

  GetTodo:
    method: GET
    path: /todos/{{ id }}

  CreateTodo:
    method: POST
    path: /todos/
    jsonBody: |
      {
        "title": "{{ title }}"
      }
```

![](./media/quick-start/quick-start-post-request.gif)

`jsonBody` automatically sets the `Content-Type` header to `application/json`.

## Enforcing types on user input variables

We can enforce types on user input variables by specifying the type following the format `<varname:type>`.

```yaml title="http.yaml" hl_lines="21"
config:
  host: jsonplaceholder.typicode.com
  port: 443
  scheme: https

requests:
  ListTodos:
    method: GET
    path: /todos

  GetTodo:
    method: GET
    path: /todos/{{ id }}

  CreateTodo:
    method: POST
    path: /todos/
    jsonBody: |
      {
        "title": "{{ title }}",
        "userId": "{{ userId:int }}
      }
```

Type of variable is displayed in the prompt.

![](./media/quick-start/quick-start-type-enforcement.gif)

If the user enters a value of a different type, the request will fail.

![](./media/quick-start/quick-start-type-enforcement-fail.gif)