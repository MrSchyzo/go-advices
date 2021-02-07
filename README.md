# go-advices

A simple jsonrpc v2.0 wrapper over [advice slip service](https://api.adviceslip.com/)

## How to run

- make sure you have a container runtime able to read multistage Dockerfiles (eg. Docker 17.05+)
- run `docker build -f Dockerfile -t go-advices:latest` in the root folder of this project;
- run `docker run --rm -d -p <aPortNumber>:10000 --name go-advices localhost/go-advices:latest` where `<aPortNumber>` can be any free local port; you can use port `10000` if it is not used;
- you can send any JSONRPC v2.0 HTTP POST messages to `localhost:10000/rpc` with method name `AdviceService.GiveMeAdvice`, `Accept` and `Content-Type` headers set to `application/json`;
- run `docker stop go-advices` when you are done with the execution of the containerized application

## TODO

- add some unit tests (maybe)
