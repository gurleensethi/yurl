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

![](./media/quick-start/quick-start-1-basic.gif)

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

![](./media/quick-start/quick-start-2-user-input.gif)

Learn more about variables [here](./variables.md).

## List Requests

We can list all the requests defined in the `http.yaml` file by running `yurl ls` command.

```bash
yurl ls
```

![](./media/quick-start/quick-start-3-list-requests.gif)