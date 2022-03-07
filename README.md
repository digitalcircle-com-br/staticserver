# staticserver
Simple server for static files

# Install

```shell
go install github.com/digitalcircle-com-br/staticserver
```

# Usage
call staticserver, it will start serving http from static folder inside present folder

# Client Side Config
To make easier to use the same project in different contexts, a config file may be forwarded to the clients, calling the __config path at static server top level.

This feature relies on the existence of a file called config.yaml, with the following format:

```yaml
"localhost:8080":
  a: "ASD"
  b: "123"
  c: 1
  d:
    da: da
    db: 23
"api.domain.com/apinst":
  a: "ASD"
  b: "123"
  c: 1
  d:
    da: da
    db: 23    
```

It will resolve which object to return as json based on host and url received for this request.

Sample output
```json
{
    "__host":"localhost:8080",
    "__key":"localhost:8080",
    "__url":"/__config",
    "a":"ASD",
    "b":"123",
    "c":1,
    "d":{
        "da":"da",
        "db":23
    }
}
```
Please note __host, __key and __url will be added to be response object, to help debugging later.

An optional object called "*" may be provided in your yaml file, to be the fallback