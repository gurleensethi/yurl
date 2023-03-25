# Examples

## Basic

**Use Case**: Making a simple login POST request.

By default `yurl` loads in a file named `http.yaml`.

```yaml title="http.yaml"
config:
  host: localhost
  port: 8000

requests:
  Login:
    path: /v1/api/auth/login
    method: POST
    jsonBody: |
      { 
          "email": "{{ email }}",
          "password": "{{ password }}" 
      }
```

```bash title="bash"
yurl Login
```
