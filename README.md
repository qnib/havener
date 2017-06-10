# havener

A proxy for Docker SWARM Services, creating an endpoint for each `stack_name/service_name` so that one does not have to remember the ports. :)

## ToDo

This project just started, so it is not very mature. 

- [ ] Address internal IP (using `tasks.<service>`) in addition to proxy docker-host load-balancer
- [ ] Allow for longer URIs, so that multiple ports can have multiple subpathes.
 `havener/http1/ -> tasks.http1:8080` and `havener/http1/admin -> tasks.http1:8081`
- [ ] Allow for dynamic updates due to SWARM events
- [ ] Create docker-image to run havener

## Deploy Test-Services

```bash
$ docker stack deploy -c docker-compose.yml havener
Creating service havener_http1
Creating service havener_http2
$
```

## Run the proxy

```bash
$ go run main.go
2017/06/10 14:45:10 Connected to docker-engine 'v17.03.1-ce'
2017/06/10 14:45:10 Add URI 'havener/http2' -> 'localhost:8088'
2017/06/10 14:45:10 Add URI 'havener/http1' -> 'localhost:8080'
2017/06/10 14:45:10 Start Listening on port '0.0.0.0:9090
```

## Query the proxy

```bash
$ curl -s 127.0.0.1:9090/havener/http1
Welcome: 10.255.0.3
$ curl -s 127.0.0.1:9090/havener/http2
Welcome: 10.255.0.3
$ 
```

