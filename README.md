[![Release](https://github.com/maroda/toadlester/actions/workflows/release.yml/badge.svg)](https://github.com/maroda/toadlester/actions/workflows/release.yml)

**Toad Lester** is a configurable service that can be used for functional testing of data endpoints.
It was designed as a simple mock endpoint source for [Monteverdi](https://github.com/maroda/monteverdi).
The Endpoints provide constantly changing 'metrics' that follow certain algorithms, like monotonic "up" or "down".

## Use

### Docker

1. Create a local `.env` (or call it whatever, there are no secret values, the app doesn't automatically read it)
2. Fill it with the variable list, see **Configure** below.
3. Run it:
```shell
docker run -d --env-file ./.env --rm --network host \\
    --name toadlester ghcr.io/maroda/toadlester:latest
```
4. Test it:
```shell
$ curl localhost:8899/metrics
Metric_exp_up: 4.4e+06
Metric_exp_down: 3.0e+08
Metric_float_up: 25.41156
Metric_float_down: 9036.15245
Metric_int_up: 384
Metric_int_down: 5060
```

### Local

1. Clone this repo and `cd` there
2. `go build .`
3. Make any configuration changes you want in `./.env`
4. Source it: `set -a;. ./.env`
5. Run it: `./toadlester`

## Endpoints

### Random Metrics

<http://localhost:8899/rand/all> returns a set of all numeric types supported that change randomly every second:
```shell
$ curl localhost:8899/rand/all
ExpMetric: 2.00028e+09
FloatMetric: 48490921.17416
IntMetric: 1036086118
```

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
- Size (`SIZE`) is the number of elements in the metric series (which are repeated). A small size is a series that fires quicker and jumps through its range faster. Increase this to get longer spans of numbers.
- Limit (`LIMIT`) is an integer for capping metric values. Increase this to get larger numbers.
- Tail (`TAIL`) is used for decimal places, integer types ignore it. For floats this is precision, for exponents this is the mantissa. 
- Mod (`MOD`) is a float used as a multiplier. Increase this with `LIMIT` to get very large numbers.

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

### Reset for New Values

Each of the configuration Env Vars can be changed while the app is running. For instance, `localhost:8899/reset/INT_SIZE/1000` changes the running `INT_SIZE` variable to `1000` and fills the buffer with a completely new set of values.
In this case, such a setting will create a series of 1000 upwards integers for the `/series/int/*` endpoints.

## Monteverdi Configuration

Compatible `config.json` for use with [Monteverdi](https://github.com/maroda/monteverdi).

```json
[
  {
    "id": "TOADLESTER_RANDOMIZER",
    "url": "http://localhost:8899/rand/all",
    "delim": ": ",
    "interval": 2,
    "metrics": {
      "ExpMetric": {
        "type": "gauge",
        "transformer": "",
        "max": 50000
      },
      "FloatMetric": {
        "type": "gauge",
        "transformer": "",
        "max": 50000
      },
      "IntMetric": {
        "type": "gauge",
        "transformer": "",
        "max": 50000
      }
    }
  },
  {
    "id": "TOADLESTER_SERIAL",
    "url": "http://localhost:8899/metrics",
    "delim": ": ",
    "interval": 3,
    "metrics": {
      "Metric_exp_down": {
        "type": "gauge",
        "transformer": "",
        "max": 1600000000
      },
      "Metric_exp_up": {
        "type": "gauge",
        "transformer": "",
        "max": 16000000
      },
      "Metric_float_down": {
        "type": "gauge",
        "transformer": "",
        "max": 500
      },
      "Metric_float_up": {
        "type": "gauge",
        "transformer": "",
        "max": 600
      },
      "Metric_int_down": {
        "type": "gauge",
        "transformer": "",
        "max": 400
      },
      "Metric_int_up": {
        "type": "gauge",
        "transformer": "",
        "max": 50
      }
    }
  }
]
```