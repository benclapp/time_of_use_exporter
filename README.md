# time_of_use_exporter

Prometheus exporter to auto generate Time of Use style metrics, in a specific timezone. PromQL natively only supports the UTC timezone, making it painful to calculate time of day based metrics, especially in custom timezones.

This exporter exposes Prometheus metrics based on configured time windows, and timezones. Some localized equivalent metrics to replace the native `minute()`, `hour()`, `day_of_week()`, `day_of_month()`, and `month()` PromQL functions are also produced.

## Config

Time of use metrics are configured with a configuration file. An annotated example:

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
  # Value to return if no time windows match
  default_value: 0.1106
  # List of time window overrides for alternate values
  # First match in the list will be used
  # List order is not guaranteed, so for certainty don't configure overlapping windows
  time_windows:
    # Override value
  - value: 0.2423
    # Start of the window, in hh:mm 24h format
    start: '7:00'
    # end of the window
    end: '9:00'
  - value: 0.19
    start: '9:00'
    end: '17:00'
  - value: 0.2423
    start: '17:00'
    end: '21:00'

```
