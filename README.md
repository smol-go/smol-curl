A minimal clone of cURL

### generate certificates required for -E flag
```
# Generate a private key
openssl genpkey -algorithm RSA -out client-key.pem

# Generate a certificate signing request (CSR)
openssl req -new -key client-key.pem -out client-req.pem

# Generate a self-signed certificate
openssl x509 -req -days 365 -in client-req.pem -signkey client-key.pem -out client-cert.pem

# Combine key and certificate into a single file
cat client-key.pem client-cert.pem > client.pem
```

### tests to try out different flags
default
```
./main https://www.keycdn.com
```

-a
```
./main -A "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36" https://www.keycdn.com
```

-e
```
./main -E client.pem https://www.keycdn.com
```

-I
```
./main -I https://www.keycdn.com
```

-k
```
./main -k https://www.keycdn.com
```

-d
```
./main -d "name=John" -d "age=30" http://example.com/form
```

-H
```
./main -H "X-Custom-Header: Value" -H "Authorization: Bearer token" http://example.com
```

-u
```
./main -u username:password http://example.com
```

-F
```
./main -F "name=John" -F "file=@/path/to/file.txt" http://example.com/upload
```
Note that if both -F and -d flags are used, the -F flag takes precedence, and the request will be sent as multipart/form-data.