# distributed_service_with_go
distributed_service_with_go study


## test code
```bash
$ curl -X POST localhost:8080 -d '{"record": {"val
ue": "TGV0J3MgR28GiZEK"}}'
{"offset":0}

$ curl -X GET localhost:8080 -d '{"offset": 0}'
{"record":{"value":"TGV0J3MgR28GiZEK","offset":0}}
```