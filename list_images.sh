#! /usr/bin/env sh
curl -X GET -H "Content-Type: application/json" -H "Authorization: Bearer `cat dospin.json | jq -r .Token`" "https://api.digitalocean.com/v2/images?page=1&per_page=100&private=true" | jq .
