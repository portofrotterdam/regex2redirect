# Regex2Redirect

Regex2Redirect is a middleware plugin for [Traefik](https://github.com/traefik/traefik) to redirect a request based on the response body.

Content-Encoding is unsupported, this means only responses without the "Content-Encoding" header or of which the value "entity" can be processed. In any other case there will be a UnprocessableEntity 422 response code.
If the regex evaluation has no match or the value can't be parsed as URL this will return a NotFound 404 response code.

## Configuration

### Static

```yaml
pilot:
  token: "xxxxx"

experimental:
  plugins:
    regex2redirect:
      moduleName: "github.com/portofrotterdam/regex2redirect"
      version: "v0.0.1"
```

### Dynamic

```yaml
http:
  middlewares:
    regex2redirect-foo:
      regex2redirect:
        regex: '\w+:(\/?\/?)[^\s"]+'
```
