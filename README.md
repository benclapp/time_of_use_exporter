# time_of_use_exporter

Prometheus exporter to auto generate Time of Use style metrics, in a specific timezone. PromQL natively only supports the UTC timezone, making it painful to calculate time of day based metrics, especially in custom timezones.

This exporter exposes Prometheus metrics based on configured time windows, and timezones. Some localized equivalent metrics to replace the native `minute()`, `hour()`, `day_of_week()`, `day_of_month()`, and `month()` PromQL functions are also produced.

## Config

Environment variables:

| variable      | description                                                           | default         |
| ------------- | --------------------------------------------------------------------- | --------------- |
| `LISTEN_ADDR` | Address to listen for metrics on                                      | `:10007`        |
| `CONFIG_FILE` | Relative path to the configuration file.                              | `./config.yaml` |
| `LOG_LEVEL`   | Log level. Accepted values are: "debug", "info", "warn", and "error". | `info`          |

Time of use metrics are configured with a configuration file. Changes to this file while the exporter is running will be automatically reloaded.

An annotated example:

```yaml
# Timezones to produce timezone specific series for.
# https://www.iana.org/time-zones
localized_timezones:
- Pacific/Auckland

# List of configs for time of use series
time_of_use:
  # Metric name
- name: electricity_price
  # Metric help
  description: Electricity price
  # Timezone to observe time of use values in. If unset, UTC is used
  timezone: Pacific/Auckland
  # Map of additional labels to add to the metric.
  # A `tz` label is also added for the configured timezone
  labels:
    provider: Power Company
    plan: Zappy
    rate: Night
  # Value to return if no time windows match
  default_value: 0.1106
  # List of time window overrides for alternate values
  # First match in the list will be used
  # List order is not guaranteed, so for certainty don't configure overlapping windows
  time_windows:
    # Override value
  - value: 0.2423
    # Start of the window, in hh:mm 24h format
    start: '07:00'
    # end of the window
    end: '09:00'
    # Map of labels that will be present within this time window
    # These will override the default labels with a matching name
    labels:
      rate: Peak
  - value: 0.19
    start: '09:00'
    end: '17:00'
    labels:
      rate: Day
  - value: 0.2423
    start: '17:00'
    end: '21:00'
    labels:
      rate: Peak
    # Days of the week the filter is valid for https://pkg.go.dev/time#Weekday
    days: [1, 2, 3, 4, 5]
  - value: 0.15
    start: '21:00'
    # Can set end as midnight by using either 00:00 or 24:00
    end: '00:00'
    labels:
      rate: Peak

```
