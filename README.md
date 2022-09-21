# request

http request library for json apis

## Installation

```bash
go get github.com/dillonstreator/request
```

## Usage

```go
client := request.NewClient("https://jsonplaceholder.typicode.com/todos")

todos := []struct {
    ID        int    `json:"id"`
    UserID    int    `json:"userId"`
    Title     string `json:"title"`
    Completed bool   `json:"completed"`
}{}

values := url.Values{}
values.Add("userId", "2")

res, err := client.Get(context.Background(), "/", values, &todos)
if err != nil {
    log.Fatal(err)
}

fmt.Println(res)
fmt.Println(todos)
```

### Custom http client

```go
httpClient := &http.Client{
    Timeout: time.Second * 5,
}

client := request.NewClient(
    "https://jsonplaceholder.typicode.com/todos",
    request.WithHTTPClient(httpClient),
)
```

### Custom error handling

```go
client := request.NewClient(
    "https://jsonplaceholder.typicode.com/todos",
    request.WithErrChecker(func(req *http.Request, res *http.Response) error {
        if res.StatusCode != http.StatusOK { // your custom error handling here...
            return fmt.Errorf("some error occurred %d %s%s", res.StatusCode, req.URL.Host, req.URL.Path)
        }

        return nil
    }),
)

items := []struct {
    // ...
}{}
_, err := client.Get(context.Background(), "/", nil, &items)
if err != nil {
    log.Fatal(err)
}

fmt.Println(items)
```

### Bearer token auth

```go
client := request.NewClient(
    "https://some-bearer-token-authed-api.com",
    request.WithBearerToken("<token-here>"),
)
```

### Basic auth

```go
client := request.NewClient(
    "https://some-basic-token-authed-api.com",
    request.WithBasicAuth("user", "pass"),
)
```

### All together

```go
customHTTPClient := &http.Client{
    Timeout: time.Second * 5,
}

client := request.NewClient(
    "https://some-bearer-token-authed-api.com",
    request.WithHTTPClient(customHTTPClient),
    request.WithBearerToken("<token-here>"),
    request.WithErrChecker(func(req *http.Request, res *http.Response) error {
        if res.StatusCode != http.StatusOK {
            return fmt.Errorf("some error occurred %d %s%s", res.StatusCode, req.URL.Host, req.URL.Path)
        }

        return nil
    }),
)
```