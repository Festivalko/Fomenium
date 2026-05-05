FROM p4lang/p4c:latest AS compiler
# Фиксим битые библиотеки
RUN apt-get update && apt-get install -y --reinstall libboost-iostreams1.71.0 && rm -rf /var/lib/apt/lists/*
WORKDIR /p4
COPY p4/aggregator.p4 .
RUN p4c-bm2-ss --p4v 16 aggregator.p4 -o aggregator.json

FROM p4lang/behavioral-model:latest
RUN apt-get update && apt-get install -y netcat-openbsd && rm -rf /var/lib/apt/lists/*
WORKDIR /p4
COPY --from=compiler /p4/aggregator.json .
EXPOSE 9090
CMD ["simple_switch", "--log-console", "--thrift-port", "9090", "aggregator.json"]
