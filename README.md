# urlShortenerAPI

- A simple RESTful API for storing URLs by their alias
- Using `gin` framework (the previous version used just the standard library)
- Links are stored in a Red-Black tree
- URLs are validated
- Safe operations with the datastructure using `sync` package
- Data stored in RAM

## Usage

### Get all alias-URL pairs
Send a `GET` request to endpoint `/links`

```bash
curl localhost:8090/links --request "GET"
```

### Get URLs with aliases with specified prefix
Send a `GET` request to endpoint `/links` and `prefix` query parameter
```bash
curl localhost:8090/links?prefix=pr --request "GET"
```

### Get a sigle URL by alias
Send a `GET` request to endpoint `/links/<alias>`
```bash
curl localhost:8090/links/myAlias --request "GET"
```

### Add new alias and URL
Send a `POST` request to endpoint `/links` with a `.json` body:
```json
{
    "alias": "ex",
    "url": "https://www.example.com"
}
```
```bash
curl localhost:8090/links --request "POST" -d @body.json --header "Content-Type: application/json" 
```

### Change url at alias
Send a `PATCH` request to endpoint `/links/<alias>` with a `.json` body:
```json
{
    "url": "https://www.other.com"
}
```
There should be no alias field in json or it should be equal to current alias.
```bash
curl localhost:8090/links/link --request "PATCH" -d @new_body.json --header "Content-Type: application/json"
```

### Delete an alias
Send a `DELETE` request to endpoint `/links/<alias>`
```bash
curl localhost:8090/links/myAlias --request "DELETE"
```