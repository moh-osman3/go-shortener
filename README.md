## URL Shortener Go Server

A scalable url shortening server with in-memory caching and durable storage.

## How to start the server
1. `git clone https://github.com/moh-osman3/go-shortener.git`
2. `cd go-shortener/`
3. `go mod tidy` to help fetch dependencies
4. `cd cmd/shortener`
5. `go run .`

This will start up a server running on localhost:3030

## How to interact with the server

The server has provides endpoints for creating a short-url, getting a short-url or its summary, and deleting a short-url.

# Creating a short url

Creating a shorturl requires a POST request that accepts data in the following format
{
    "url": <long-url-string>,
    "expiry": <time-duration-string>
}

"url": takes in a long url that you want to encode
"expiry": takes in a time duration (e.g. "10s", "10m", "10d", etc). 0 or unset expiry defaults to 1 year expiration date. Negative expiry (e.g. "-10s") means the short url should have no expiration date.

Putting it all together, the following request will create a shorturl for www.google.com with no expiration date.

`curl -X POST -d '{"url":"www.google.com","expiry":"-3s"}' http://localhost:3030/create`

# Getting a short url

Fetching your short url to redirect to the long url requires a GET request to the server.

The endpoint is http://localhost:3030/<short-url-hash>

e.g. to redirect to google.com using the shorturl

`curl http://localhost:3030/MA==`

# Getting a summary of your short url

The server supports instrumentation that records the number of times the short url has been called in the past day, the past week, and all time.

The endpoint is http://localhost:3030/<short-url-hash>/summary

So to get a summary of your google.com shorturl you can write
`curl http://localhost:3030/MA==/summary`

```
Summary of shorturl:
 calls in the last day: 1 calls
 calls in the last week: 1 calls
 total calls since creation: 1 calls
```

# Deleting a short url

To delete a short url the server expects a DELETE request that accepts data in the following format

{
    "url": <short-url-hash>,
}

Putting it all together, the following request will delete the short url for www.google.com

`curl -X DELETE -d '{"id":"MA=="}' http://localhost:3030/delete`

## Testing

To run tests on the source code go to the root of the repository and run `go run ./... -v -race`

Current code coverage:
```
github.com/moh-osman3/shortener              coverage: 66.7% of statements
github.com/moh-osman3/shortener/managers/def coverage: 83.5% of statements
github.com/moh-osman3/shortener/urls         coverage: 96.1% of statements
```

## Design

Features:
- creating a shorturl from a long url
- redirecting a short url to a long url
- deleting short urls
- summary of number of redirects for a shorturl
- data persistence
- logging to help with debugging

# URL generation

Short Url generation uses sequential ID's and apply a one-to-one feistel transformation to get a unique obfuscated encoding. Then we apply a url safe encoding to the transformation for our short url. A long url should only map to one shortUrl for the lifetime of that shortUrl. If the shortUrl is deleted, then the next time the long url is submitted, it will generate a new unique shortUrl.

# Storage

This server supports two layer storage with an in memory cache and leveldb for durable storage. The in memory cache helps performance for accessing frequently used short urls and deleting expired keys within the cache. Background threads will periodically scan keys in the cache or db to find and delete expired keys. Consistency is ensured by having any operation on the cache be reflected in the db as a single transaction and vice versa. This has a performance cost which could be improved by sacrificing consistency and batching db writes. 
LevelDB was selected for its simplicity because we are storing relatively simple key-value pairs with no complex relationships. The leveldb client was preferred over something like redisdb because leveldb writes directly to the local file system for persistance while redisdb depends on the redis server and external configs for durability.

# Expiration

Short urls have an optional expiration date measured as a golang time.Duration. For users that don’t provide an expiration date, the expiration will default to 365 days. This ensures we don’t perpetually store unused short URLs and frees up space in the db. We can have a background thread that scans our db or cache layer for expired shortUrls. In the future we will support the ability to configure the periodicity of the cache and db cleanups. A user can specify no expiration date by specifying a negative duration. A 0 duration will be interpreted the same as an unset duration and default to 365 days.

# Instrumentation

The server supports a very inefficient method of counting the number of calls to a short url in the past day, week and all time. The counter stores a map of days to number of calls, scans all the keys and aggregates the sum for daily, weekly, and all time totals. In the future we can use otel instrumentation or create our own custom histogram to store our summary for the past 7 days.

We can also add prometheus metrics to emit these metrics at port 8888 for other applications to scrape.
