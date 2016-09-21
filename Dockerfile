FROM alpine:3.2
ADD prometheus-cloudstack-discovery /bin/prometheus-cloudstack-discovery
ENTRYPOINT ["/bin/prometheus-cloudstack-discovery"]
