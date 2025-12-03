**Toad Lester** is a configurable service that can be used for functional testing of data endpoints.

## Use

> Package coming soon

1. Clone this repo and `cd` there
2. `go build .`
3. Make any configuration changes you want in `./.env`
4. Source it: `set -a;. ./.env`
5. Run it: `./toadlester`



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

#### All At Once

<http://localhost:8899/metrics> is a full page of each metric. These change every second as an internal process updates metrics using each algorithm.
```shell
$ curl localhost:8899/metrics
Metric_exp_up: 1.81950755e+07
Metric_exp_down: 1.57610423e+09
Metric_float_up: 181.41967953
Metric_float_down: 318.81474780
Metric_int_up: 48
Metric_int_down: 480
```

#### Series API
<http://localhost:8899/series/> is an API endpoint that can be configured with a type and an algorithm.

For instance, <http://localhost:8899/series/int/up> requests the integer series that is being advanced monotonically upward, changing every second.
```shell
$ curl localhost:8899/series/int/up
Metric_int_up: 2
$ curl localhost:8899/series/int/up
Metric_int_up: 3
```

## Configure

The configuration defines things like the digits of the number and how many times it rises. Once the series reaches the end, it cycles and starts from the beginning.

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

This is a working config example:
```dotenv
EXP_SIZE=5
EXP_LIMIT=250
EXP_TAIL=1
EXP_MOD=250
FLOAT_SIZE=4
FLOAT_LIMIT=100
FLOAT_TAIL=5
FLOAT_MOD=1.123
INT_SIZE=10
INT_LIMIT=100
INT_TAIL=1
INT_MOD=1
RAND_SIZE=1
RAND_LIMIT=500
RAND_TAIL=3
RAND_MOD=500
```

## TODO

1. Temporal interval of metric - how much time between events - should be configurable. Right now, these are simple sinusoidal waveforms of metrics.
2. Reset control to re-read configurations and/or have a way to update them while running

