#!/bin/bash

#curl -X 'PUT' \
#  'http://10.67.100.103:32265/api/v1/provisioning/policies' \
#  -H 'accept: application/json' \
#  -H 'Content-Type: application/json' \
#  -H 'cookie: grafana_session=5ec944b435575623779c846f04003d21' \
#  -d '{
#{
#  "receiver": "webhook",
#  "group_by": [
#    "alertname"
#  ],
#  "group_wait": "30s",
#  "group_interval": "5m",
#  "repeat_interval": "4h"
#}'

curl -X 'PUT' \
  'http://10.67.100.103:32265/api/v1/provisioning/policies' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "active_time_intervals": [
    "string"
  ],
  "continue": true,
  "group_by": [
    "string"
  ],
  "group_interval": "string",
  "group_wait": "string",
  "match": {
    "additionalProp1": "string",
    "additionalProp2": "string",
    "additionalProp3": "string"
  },
  "match_re": {
    "additionalProp1": "string",
    "additionalProp2": "string",
    "additionalProp3": "string"
  },
  "matchers": [
    {
      "Name": "string",
      "Type": 0,
      "Value": "string"
    }
  ],
  "mute_time_intervals": [
    "string"
  ],
  "object_matchers": [
    [
      "string"
    ]
  ],
  "provenance": "string",
  "receiver": "string",
  "repeat_interval": "string",
  "routes": [
    "string"
  ]
}'
