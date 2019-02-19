This is a simple microservice to track cryptocurrencies exchange rates.
The service consists of 2 applications and a database, all as docker images:

1. data_collector: crawls public apis for market information.
   configure the currency pairs of interest by the environment variable `MARKETS`

2. api: the public API offering information on markets of interests, currently
   offering this endpoint:
   `/markets/<market-name>?from=...&to=...`
   `<market-name>` is optional, when omitted, it returns all markets available
   `from` and `to` are datetime in RFC3339 format
   example:
   /markets/ETH-ADA?from=2019-02-19T10:59:00Z&to=2019-02-19T11:01:21Z

   response:
   ```
   [
        {
            "from": "2019-02-19T11:01:14Z",
            "to": "2019-02-19T11:01:19Z",
            "data": [
            {
                "market": "ETH-ADA",
                "low": 0.0003029,
                "high": 0.00032621,
                "volume": 1618015.10968697
            }
            ]
        },
        ...
    ]
  ```

Both applications are pointed to the mongodb server by the environment variable `MONGO_ADDRESS`

To start the up:

`docker-compose build`
`docker-compose up`


To test the app, make sure you have pytest and datadiff installed, e.g.:
`pip install pytest datadiff`

Set the database to acceptance state:
run: service
	docker-compose build
	docker-compose up

test: test-acceptance

test-acceptance: tests/acceptance
    mongoimport --host localhost --port 5000 \
        --db exchange --collection markets --drop \
        --file tests/data/acceptance.json --jsonArray


Then run:
`pytest tests/acceptance`

Improvements:
- modular database access package
- unit and integration tests


