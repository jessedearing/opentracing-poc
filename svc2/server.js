const express = require('express')
const Zipkin = require('zipkin')
const app = express()
const CLSContext = require('zipkin-context-cls')
const hLogger = require('zipkin-transport-http')
const ctxImpl = new CLSContext()
const xtxImpl = new Zipkin.ExplicitContext()
//const recorder = new Zipkin.BatchRecorder({
  //logger: new hLogger.HttpLogger({
    //endpoint: 'http://10.3.0.39:9411/api/v1/spans'
  //})
//});
const recorder = new Zipkin.BatchRecorder({
  logger: new hLogger.HttpLogger({
    endpoint: 'http://localhost:9411/api/v1/spans'
  })
});
const zipkinMiddleware = require('zipkin-instrumentation-express').expressMiddleware

const tracer = new Zipkin.Tracer({
  ctxImpl,
  recorder: recorder,
});

app.use(zipkinMiddleware({
  tracer,
  serviceName: 'service2'
}));

app.get("/", function(req, res) {
  res.send("Hi there")
});

app.get("/api/slow", function(req, res) {
  setTimeout(function() {
    res.send({ hello: 'world' })
  }.bind(this), 1400)
})

app.listen(3000, function() {
  console.log("Application started")
})
