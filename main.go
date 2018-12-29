package main

import (
	"log"
	"net/http"
	"os"

	"github.com/stuart-warren/serveit/access"
	"github.com/stuart-warren/serveit/middleware"
	"github.com/stuart-warren/serveit/middleware/logging"
	"github.com/stuart-warren/serveit/middleware/metrics"
	"github.com/stuart-warren/serveit/router"
	"github.com/stuart-warren/serveit/rules"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	wd, _ := os.Getwd()
	static := http.FileServer(http.Dir(wd))
	metricMux := router.NewRouter(promhttp.Handler(), rules.AllowAll)
	metricMux.Handle(router.NewPrefixRoute("/metrics").Permit(access.BlankPermit().MethodRO().AllowUsers("ALL")))
	phm := metrics.NewPrometheusHttpMetric("serveit", []float64{50.0, 90.0, 95.0, 99.0, 99.999})
	rootMux := router.NewRouter(static, rules.CheckUser)
	rootMux.Handle(router.NewPrefixRoute("/access/").Permit(access.BlankPermit().MethodRW().AllowUsers("some.admin")))
	rootMux.Handle(router.NewPrefixRoute("/").Permit(access.BlankPermit().MethodRO().AllowUsers("ALL")))
	mux := http.NewServeMux()
	mux.Handle("/metrics", metricMux)
	mux.Handle("/", rootMux)
	srv := &http.Server{Addr: ":1234", Handler: middleware.Decorate(mux, phm.For("/"), logging.Handler())}
	log.Printf("starting at %s\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
