package controller

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/appscode/pat"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

func (c *Controller) runHTTPServer() {
	m := pat.New()
	m.Get("/metrics", promhttp.Handler())
	http.Handle("/", m)

	log.Infof("Starting Server: %s", c.opt.Address)
	log.Fatal(http.ListenAndServe(c.opt.Address, nil))
}
