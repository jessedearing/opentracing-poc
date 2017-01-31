package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	log "github.com/sirupsen/logrus"
)

type HelloData struct {
	Hello string
}

func main() {
	//collector, err := zipkin.NewHTTPCollector("http://10.3.0.39:9411/api/v1/spans")
	collector, err := zipkin.NewHTTPCollector("http://localhost:9411/api/v1/spans")
	if err != nil {
		log.Fatal(err)
	}

	tracer, err := zipkin.NewTracer(
		zipkin.NewRecorder(collector, false, "172.0.0.1:0", "service1"))

	if err != nil {
		log.Fatal(err)
	}

	opentracing.InitGlobalTracer(tracer)

	http.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		initTracer := opentracing.StartSpan("frontend request")
		defer initTracer.Finish()

		resp := callSlow("slow1", initTracer)
		resp.Body.Close()
		resp = callSlow("slow2", initTracer)
		defer resp.Body.Close()

		h := &HelloData{}

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			return
		}

		err = json.Unmarshal([]byte(bytes), h)
		if err != nil {
			log.Error(err)
			return
		}

		response.Write([]byte(fmt.Sprintln(h)))

	})

	http.HandleFunc("/shutdown", func(response http.ResponseWriter, request *http.Request) {
		time.Sleep(2 * time.Second)
		response.Write([]byte{})
		os.Exit(0)
	})

	http.HandleFunc("/favicon.ico", func(response http.ResponseWriter, request *http.Request) {
		response.Write([]byte{})
	})

	http.ListenAndServe(":3001", nil)

	time.Sleep(2 * time.Second)
}

func callSlow(name string, rootTrace opentracing.Span) http.Response {
	httpClient := &http.Client{}
	clientReq, err := http.NewRequest("GET", "http://localhost:3000/api/slow", nil)
	if err != nil {
		log.Error(err)
		return http.Response{}
	}
	span := opentracing.StartSpan(name, opentracing.ChildOf(rootTrace.Context()))
	carrier := opentracing.HTTPHeadersCarrier(clientReq.Header)

	err = span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier)
	if err != nil {
		log.Error(err)
		return http.Response{}
	}

	resp, err := httpClient.Do(clientReq)
	if err != nil {
		log.Error(err)
	}
	span.Finish()

	return *resp
}
