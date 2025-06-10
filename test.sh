curl -X 'POST' \
	'http://10.67.100.103:8080/api/v1/simplealert' \
	-H 'accept: application/json' \
	-d '{
            "cpu": "2",
            "memory": "50",
            "disk": "50"
	     }'
