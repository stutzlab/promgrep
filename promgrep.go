package promgrep

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type PromOptions struct {
	bindHost    string
	bindPort    uint
	metricsPath string
}

type MetricRule struct {
	Name   string
	Regex  string
	metric prometheus.Counter
	in     chan string
}

func Start(ctx context.Context, rules []MetricRule, opt PromOptions, in io.Reader) error {
	if opt.bindHost == "" {
		opt.bindHost = "0.0.0.0"
	}
	if opt.bindPort == 0 {
		opt.bindPort = 8880
	}
	if opt.metricsPath == "" {
		opt.metricsPath = "/metrics"
	}

	logrus.Debugf("Setup prometheus metrics...")
	router := mux.NewRouter()
	router.Handle(opt.metricsPath, promhttp.Handler())

	r2 := make([]MetricRule, 0)
	for _, r := range rules {
		r.metric = promauto.NewCounter(prometheus.CounterOpts{
			Name: fmt.Sprintf("promgrep_%s", r.Name),
			Help: fmt.Sprintf("Counter for regex '%s'", r.Regex),
		})
		r.in = make(chan string)
		logrus.Debugf("Preparing rule %s", r.Name)
		r2 = append(r2, r)
	}

	listen := fmt.Sprintf("%s:%d", opt.bindHost, opt.bindPort)
	listenPort, err := net.Listen("tcp", listen)
	if err != nil {
		panic(err)
	}

	go func() {
		logrus.Infof("Starting Prometheus Exporter at http://%s:%d%s\n", opt.bindHost, opt.bindPort, opt.metricsPath)
		http.Serve(listenPort, router)
		defer listenPort.Close()
	}()
	go func() {
		select {
		case <-ctx.Done():
			listenPort.Close()
		}
	}()

	logrus.Debugf("Preparing stream reader...")
	go func(rs []MetricRule) {
		scanner := bufio.NewScanner(in)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				break
			default:
				for _, r := range rs {
					r.in <- scanner.Text()
				}
			}
		}
		for _, r := range rs {
			close(r.in)
		}
	}(r2)

	logrus.Debugf("Preparing regex scanners...")
	for _, rul := range r2 {
		reg := regexp.MustCompile(rul.Regex)
		go func(ru MetricRule) {
			logrus.Debugf("Routine for '%s' started", ru.Name)
			defer func() {
				logrus.Debugf("Routine for '%s' stopped", ru.Name)
			}()
			for line := range ru.in {
				select {
				case <-ctx.Done():
					return
				default:
					matches := reg.FindAllStringSubmatch(line, 99)
					if matches == nil {
						continue
					}
					for _, m := range matches {
						c := 1.0
						var err2 error
						if len(m) > 1 {
							c, err2 = strconv.ParseFloat(m[1], 64)
							if err2 != nil {
								logrus.Warnf("Could not parse '%s' as float in stream", m[1])
								continue
							}
						}
						ru.metric.Add(c)
					}
				}
			}
		}(rul)
	}

	return nil
}
