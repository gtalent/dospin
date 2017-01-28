#! /usr/bin/env sh
curl -X GET -H "Content-Type: application/json" -H "Authorization: Bearer `cat dospin.json | jq -r .ApiToken`" "https://api.digitalocean.com/v2/droplets?page=1&per_page=100&private=true" | jq .
