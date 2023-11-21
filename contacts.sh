#!/usr/bin/env bash


head contacts.csv | awk 'BEGIN { FS = "," } ; { print $1","$2","tolower($1) }' | while IFS=',' read name addr namel; do; k apply -f - <<EOF
apiVersion: hiring.influxdata.io/v1alpha1
kind: EmailRequest
metadata:
  name: emailreq-${namel// /.}
spec:
  name: $name
  address: $addr
  retryBlockedPolicy: true
EOF
;done;
