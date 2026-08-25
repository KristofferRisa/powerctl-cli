# Powerctl Domain

This context covers the core terminology for the Tibber API integration and the powerctl-cli functionality.

## Language

**Consumption History**:
A record of power used by a home over a specific time period, including the total consumption (kWh), cost, and unit prices.
_Avoid_: Invoices, billing statements, usage

**Resolution**:
The time grouping for consumption data, mapping directly to Tibber's API (hourly, daily, weekly, monthly, annual).
_Avoid_: Granularity, interval, bucket

**Node**:
A single data point in the consumption history representing a specific period (from/to) and its associated metrics.
_Avoid_: Point, row, entry
