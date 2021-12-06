# prom2mdi
Prometheus(AlertManager) to MDI Service

```flow
pro=>start: Prometheus
mdi=>end: MDI Serivce
alert=>operation: AlertManger
hook=>operation: prom2md

pro->alert->hook->mdi
```
