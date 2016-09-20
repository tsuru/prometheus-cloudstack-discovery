Prometheus CloudStack Service Discovery
=======================================

`prometheus-cloudstack-discovery` generates target groups by querying the
CloudSTack API. This is to be used in conjunction with Prometheus's
[custom service discovery][1] feature.

Example
-------

```
$ prometheus-cloudstack-discovery
[
  {
    "targets": ["172.22.2.57:9100"],
    "labels": {"job": "cadvisor"}
  }
]
```
Install
-------

First install the binary

```
$ go get github.com/tsuru/prometheus-cloudstack-discovery
```

Then configure a [file based service discovery] [1] target group in your
`prometheus.yml` config:

```yaml
scrape_configs:
  - job_name: my_cloudstack_service

    file_sd_configs:
    - names: ['tgroups/*.json']
```

You can then setup a crontab job to update the target groups every minute:

```
* * * * * cd /path/to/prometheus/dir && prometheus-cloudstack-discovery --dest=tgroups/cloudstack.json
```

[1]: http://prometheus.io/blog/2015/06/01/advanced-service-discovery/#custom-service-discovery "Custom Service Discovery"
