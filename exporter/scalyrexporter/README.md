# Scalyr Exporter
 
Exports trace data to a [Scalyr](https://scalyr.com/)

The following settings are required:

- `endpoint` (default = https://app.scalyr.com): Scalyr Endpoint to send trace data.
- `api_key` (No default): Scalyr API Key for authentication

You can also set environment variables:

SCALYR_ENDPOINT
SCALYR_API_TOKEN

Example:

```yaml
exporters:
scalyr:
  api_key: 123456789
```
