# Count

Increment and decrement a common counter.

## Running the Example

`https` is required for `http2` to work, so we need to create some certs.

```bash
mkcert -install -cert-file ./cert.pem -key-file ./key.pem localhost
```

## Usage 

### Read Count 

```bash
curl -N -s -H "x-stream: true" https://localhost:3000/count | jq
```

### Increment Count

```bash
curl -X POST https://localhost:3000/count/increment
```

### Decrement Count

```bash
curl -X POST https://localhost:3000/count/decrement
```







