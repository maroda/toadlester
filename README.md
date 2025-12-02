**Toad Lester** is a configurable service that can be used for functional testing of data endpoints.

## Endpoints

### Random Metrics

<http://localhost:8899/rand/all> returns a set of all three numeric types supported:
```shell
$ curl localhost:8899/rand/all
ExpMetric: 2.00028e+09
FloatMetric: 48490921.17416
IntMetric: 1036086118
```

These are updated by the app every second and use the `RAND_` family of environment variables to configure.

### Series Metrics

<http://localhost:8899/series/> is an API endpoint that can be configured with a type and an algorithm: <http://localhost:8899/series/int/up> requests the integer series that is being advanced monotonically upward. Like the Random Metrics, these are updated by the app every second, once for each type and algorithm.

The configuration defines things like the digits of the number and how many times it rises. Once the series reaches the end, it cycles and starts from the beginning.

Current algorithms supported by all three types (`exp`, `float`, `int`):

- `up` / `down` use integers only
- `floatup` / `floatdown` use floats, but can be integers

A steady flow of upward metrics looks like:
```shell
$ curl localhost:8899/series/int/up
Metric_int: 2
$ curl localhost:8899/series/int/up
Metric_int: 3
$ curl localhost:8899/series/int/up
Metric_int: 4
$ curl localhost:8899/series/int/up
Metric_int: 6
```

## Configure

Put configuration pairs in `.env` and set them in the environment before running (it is not read automatically), e.g.: `set -a;. ./.env`

The part before `_` is the numeric type:
- Exponent (`EXP`)
- Float (`FLOAT`)
- Integer (`INT`)
- Randomizer (`RAND`)

After the `_` is the configuration for that type:
- Size (`SIZE`) is the number of elements in the metric series (which are repeated)
- Limit (`LIMIT`) is for capping the range of each metric in the series
- Tail (`TAIL`) is used for decimal place precision
- Mod (`MOD`) applies a multiplier to the metric


```dotenv
EXP_SIZE=10
EXP_LIMIT=10
EXP_TAIL=1
EXP_MOD=1.1
FLOAT_SIZE=10
FLOAT_LIMIT=10
FLOAT_TAIL=1
FLOAT_MOD=1.1
INT_SIZE=10
INT_LIMIT=10
INT_TAIL=1
INT_MOD=1.1
RAND_SIZE=5
RAND_LIMIT=50000
RAND_TAIL=5
RAND_MOD=50000
```

## TODO

1. Temporal interval of metric - how much time between events - should be configurable. Right now, these are simple sinusoidal waveforms of metrics.

