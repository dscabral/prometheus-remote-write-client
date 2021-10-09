# prometheus-remote-write-client

You can use this api to convert a json for a prometheus particle and remote write to prometheus

Use Basic Authorization with user and password (User and Key generated on Prometheus) 

Destination URL as Param on request

`
curl --location --request POST 'localhost:8080/prom/push?url=https://prometheus.net/api/prom/push' \
--header 'Content-Type: application/json' \
--header 'Authorization: Basic dXNlcm5hbWU6cGFzc293cmQ=' \
--data-raw '[
    {
        "label": "dns_rates_total",
        "name": "P50",
        "value": 0
    }
]'`

Json Input formt 

```
[
    {
        "label": "dns_rates_total",
        "name": "P50",
        "value": 10
    }
]
