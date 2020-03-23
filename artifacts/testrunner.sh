#!/bin/ash

if [ $# -eq 0 ]; then
     ./gopherciser execute -c /etc/config-volume/testjob.json --metrics METRICPORT --logformat combined
else
	$*
fi
