#! /usr/bin/env sh

echo Volumes
curl -X GET -H "Content-Type: application/json" -H "Authorization: Bearer `cat dospin.json | jq -r .ApiToken`" "https://api.digitalocean.com/v2/volumes" | jq .
