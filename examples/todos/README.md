# Todos 

Collaborate on a single todo list

## Running the Example

`https` is required for `http2` to work, so we need to create some certs.

```bash
mkcert -install -cert-file ./cert.pem -key-file ./key.pem localhost
```

## Usage

### Read Todos 

With data updates:

```bash
curl -N -s https://localhost:3000/todos | jq
```

With json patch updates:

```bash
curl -N -s https://localhost:3000/todos?patch=true | jq 
```

### Create Todo

```bash
curl -X POST -H "Content-Type: application/json" -d '{"title":"hello world","description":"this is the decription"}' https://localhost:3000/todos 
```

### Update Todo

```bash
curl -X PUT -H "Content-Type: application/json" -d '{"title":"hello world","description":"this is the decription","checked":true}' https://localhost:3000/todos/<uuid> 
```

### Delete Todo

```bash
curl -X DELETE https://localhost:3000/todos/<uuid> 
```