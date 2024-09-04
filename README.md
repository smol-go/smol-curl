A minimal clone of cURL

### Installation

To build the program, use the following command:
```
go build src/main.go
```

To run the program, use the following syntax:
```
./main [options] <URL>
```

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

### Examples

1. `-a` `<string>`: Specify the User-Agent string
```
./simple_http_client -a "MyCustomUserAgent/1.0" http://example.com
```

2. `-k` `<bool>`: Allow insecure server connections when using SSL
```
./simple_http_client -k https://self-signed.badssl.com/
```

3. `-v` `<bool>`: Make the request more detailed
```
./simple_http_client -v http://example.com
```

4. `-m` `<int>`: Maximum time allowed for the operation in seconds
```
./simple_http_client -m 10 http://example.com
```

5. `-u` `<string>`: Specify the user name and password for server authentication
```
./simple_http_client -u "username:password" http://example.com/protected
```

6. `-d` `<string>`: HTTP POST data
```
./simple_http_client -d "name=JohnDoe&age=30" http://example.com/form-submit
```

7. `-o` `<string>`: Write the response body to the specified file
```
./simple_http_client -o "output.html" http://example.com
```

8. `-I` `<bool>`: Send HTTP HEAD request instead of GET
```
./simple_http_client -I http://example.com
```

9. `-E` `<string>`: Specify the client certificate file for HTTPS
```
./simple_http_client -E "client-cert.pem" https://example.com
```

10. `-D` `<string>`: Write the response headers to the specified file
```
./simple_http_client -D "headers.txt" http://example.com
```

11. `-X` `<string>`: Specify custom request method
```
./simple_http_client -X "DELETE" http://example.com/resource/123
```

12. `-H` `<string[]>`: Pass custom header(s) to server
```
./simple_http_client -H "X-Custom-Header: value" -H "Another-Header: another-value" http://example.com
```

13. `-F` `<string[]>`: Specify HTTP multipart POST data
```
./simple_http_client -F "field1=value1" -F "field2=@/path/to/file" http://example.com/upload
```

14. `--cookie` `<string>`: Send the specified cookies with the request
```
./simple_http_client --cookie "sessionId=abc123" http://example.com
```

15. `--connect-timeout` `<int>`: Maximum time allowed for the connection to be established in seconds
```
./simple_http_client --connect-timeout 5 http://example.com
```

### Note:
- The client supports both HTTP and HTTPS requests.
- The program will automatically attempt to use TLS when making requests to https:// URLs.
- The -k flag is useful when dealing with servers that have self-signed or otherwise invalid SSL certificates.
- The -D and -o flags can be used together to save both headers and body separately.
- If both -F and -d flags are used, the -F flag takes precedence, and the request will be sent as multipart/form-data.