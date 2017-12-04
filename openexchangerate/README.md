# Open Exchange Rate

A container to request and serve currency exchange rate data from [openexchangerates.org](https://openexchangerates.org).

Only one request sent per hour to remain under the free plan limit of 1000.

```shell
docker run -ti --rm -p 80:80 aduermael/openexchangerate APP_ID
```

Convert a given amount from one currency to another:

```
GET http://localhost:80/convert?v=100&from=EUR&to=USD
```

 