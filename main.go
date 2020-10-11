package main

import (
	"bytes"
	"fmt"
	"github.com/elastic/go-elasticsearch"
	"github.com/julienschmidt/httprouter"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"log"
	"net/http"
	"strings"
)

var logger *zap.SugaredLogger

var es *elasticsearch.Client

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_, _ = fmt.Fprint(w, "Go to /logs/:namespace/:pod/:container to get some logs\n")
}

func versionHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "version: %s (%s)", semVer, gitVersion)
}

func healthHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, _ = w.Write([]byte("OK"))
}

func logHandler(writer http.ResponseWriter, request *http.Request, ps httprouter.Params) {
	namespace := ps.ByName("namespace") // eg. tekton-pipelines
	pod := ps.ByName("pod")             // eg. 55c57500-0b79-11eb-99b4-b68ed46f9db7-godeploy-app-dt6n2-p-2g9q8
	container := ps.ByName("container") // eg. step-upgrade-environment

	logger.Infow("getting logs", "namespace", namespace, "pod", pod, "container", container)
	//request payload:
	//{"search_type":"query_then_fetch","ignore_unavailable":true,"index":["kubernetes-application-*-2020.10.10*","kubernetes-application-*-2020.10.11*"]}
	//{"size":500,"query":{"bool":{"filter":[{"range":{"@timestamp":{"gte":1602370882319,"lte":1602392482319,"format":"epoch_millis"}}},{"query_string":{"analyze_wildcard":true,"query":"+kubernetes.namespace: \"tekton-pipelines\" +kubernetes.pod.name: \"55c57500-0b79-11eb-99b4-b68ed46f9db7-godeploy-app-dt6n2-p-2g9q8\" +kubernetes.container.name: \"step-upgrade-environment\""}}]}},"sort":{"@timestamp":{"order":"desc","unmapped_type":"boolean"}},"script_fields":{},"aggs":{"2":{"date_histogram":{"interval":"15s","field":"@timestamp","min_doc_count":0,"extended_bounds":{"min":1602370882319,"max":1602392482319},"format":"epoch_millis"},"aggs":{}}}}

	query := strings.NewReader(`{
	  "query": {
		"bool": {
		  "must": [
			{ "term": { "kubernetes.namespace": "` + namespace + `" } },
			{ "term": { "kubernetes.pod.name": "` + pod + `" } },
			{ "term": { "kubernetes.container.name": "` + container + `" } }
		  ]
		}
	  }
	}`)

	rawResponse, err := es.Search(
		es.Search.WithIndex("kubernetes-application-*"),
		es.Search.WithBody(query),
		es.Search.WithSize(10000),
		es.Search.WithSort("log.offset:asc"),
		es.Search.WithPretty(),
	)
	if err != nil {
		logger.Error(fmt.Sprintf("Error querying es: %s", err))
		http.Error(writer, "unable to query ES", http.StatusInternalServerError)
		return
	}
	// build es query to get logs
	// get logs and sort, and return stream
	// convert raw response to only get
	var b bytes.Buffer
	_, serializeError := b.ReadFrom(rawResponse.Body)
	if serializeError != nil {
		logger.Error(fmt.Sprintf("Error deserializing response: %s", err))
		http.Error(writer, "unable to read ES response", http.StatusInternalServerError)
		return
	}
	data := b.Bytes()
	responseHits := gjson.GetBytes(data, "hits.total.value").Int()
	if responseHits == 0 {
		_, _ = writer.Write([]byte(""))
		return
	}

	var logData bytes.Buffer
	rawMessages := gjson.GetBytes(data, "hits.hits.#._source.message")
	if !rawMessages.IsArray() {
		logger.Panic("This shouldn't be possible")
	}
	messages := rawMessages.Array()
	logger.Info(fmt.Sprintf("found %v log lines", len(messages)))
	for _, message := range messages {
		logData.WriteString(fmt.Sprintf("%s\n", message.String()))
	}
	fmt.Fprint(writer, logData.String())
	return
}

func main() {
	router := httprouter.New()
	baseLogger, loggerError := zap.NewProduction()
	if loggerError != nil {
		log.Fatalf("can't initialize zap logger: %v", loggerError)
	}
	defer baseLogger.Sync() // flushes buffer, if any
	logger = baseLogger.Sugar()
	// Initialize a client with the default settings.
	//
	// An `ELASTICSEARCH_URL` environment variable will be used when exported.
	//
	var clientError error
	es, clientError = elasticsearch.NewDefaultClient()

	if clientError != nil {
		logger.Fatal("Error initializing %s", clientError)
		return
	}

	router.GET("/", index)
	router.GET("/version", versionHandler)
	router.GET("/healthz", healthHandler)
	router.GET("/logs/:namespace/:pod/:container", logHandler)

	logger.Info("starting up!")

	logger.Fatal(http.ListenAndServe(":8080", router))
}
