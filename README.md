# tekton-es-logs

This is an implementation of a simple log server for tekton dashboard to serve up container level logs when using elastic search as log store.

## usage:

`docker run -e "ELASTICSEARCH_URL=somees.domain.com" -P moustafab/tekton-es-logs:main`

See the tekton-dashboard documentation for more details [here](https://github.com/tektoncd/dashboard/blob/master/docs/walkthrough/walkthrough-logs.md#setting-up-the-dashboard-logs-fallback)

`curl <domain:port>/logs/:namespace/:pod/:container` for a container's logs
``

## limitations:

This is a simple server that fulfills the external log server necessary for log fallback.

This requires your logs to be parsed from standard out using the kubernetes plugin for the elastic stack. 

It assumes:

1. log offset for ordering
1. < 10000 logs
1. ES index pattern matches `kubernetes-application-*`
1. logs have kubernetes metadata from kubernetes plugin for elastic


## deploy


