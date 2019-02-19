This is a simple microservice to track cryptocurrencies exchange rates.
The service consists of 2 applications and a database, all as docker images:

1. data_collector: crawls public apis for market information.
   configure the currency pairs of interest by the environment variable `MARKETS`

2. api: the public API offering information on markets of interests, currently
   offering this endpoint:
   `/markets/<market-name>`

   example:
   /markets/ETH-ADA

   response:
   ```
    {
        "name": "ETH-ADA",
        "low": 0.00035006,
        "high": 0.000364,
        "volume": 871176.73089561
    }
  ```

Both applications are pointed to the mongodb server by the environment variable `MONGO_ADDRESS`

To start the up:

`docker-compose build`
`docker-compose up`


To test the app, make sure you have pytest installed, e.g.:
`pip install pytest`

Then run:
`pytest tests/acceptance`

Improvements:
- modular database access package
- unit and integration tests


