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
	"strings"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

//PromOptions prometheus exporter config
type PromOptions struct {
	BindHost    string
	BindPort    uint
	MetricsPath string
}

//MetricType type of metric (counter summary or gauge)
type MetricType string

const (
	//TypeSummary each sample will be summed up to a counter and another metric will count the number of samples presented. usefull for counting latencies and throughput
	TypeSummary MetricType = "TypeSummary"
	//TypeGauge each sample is set to the metric so that it can go up or down
	TypeGauge MetricType = "TypeGauge"
)

//MetricRule metric regex configuration for greping metric data from stream
type MetricRule struct {
	Name          string
	Regex         string
	Typ           MetricType
	metricSummary *prometheus.SummaryVec
	metricGauge   *prometheus.GaugeVec
	in            chan string
}

//Start initializes regex rules, starts prometheus exporter http endpoint and starts consuming reader and updating metrics
//use Context in order to wait or stop this routine
func Start(ctx context.Context, rules []MetricRule, opt PromOptions, in io.Reader) error {
	if opt.BindHost == "" {
		opt.BindHost = "0.0.0.0"
	}
	if opt.BindPort == 0 {
		opt.BindPort = 8880
	}
	if opt.MetricsPath == "" {
		opt.MetricsPath = "/metrics"
	}

	logrus.Debugf("Setup prometheus metrics...")
	router := mux.NewRouter()
	router.Handle(opt.MetricsPath, promhttp.Handler())

	r2 := make([]MetricRule, 0)
	for _, r := range rules {
		if r.Typ == "" {
			return fmt.Errorf("Typ must be defined in metric %s", r.Name)
		}
		if r.Typ == TypeSummary {
			r.metricSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
				Name: fmt.Sprintf("promgrep_%s", r.Name),
				Help: fmt.Sprintf("Counters for regex '%s'", r.Regex),
			}, []string{
				"label",
			})
		} else if r.Typ == TypeGauge {
			r.metricGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
				Name: fmt.Sprintf("promgrep_%s", r.Name),
				Help: fmt.Sprintf("Gauge for regex '%s'", r.Regex),
			}, []string{
				"label",
			})
			if (strings.Index(r.Regex, "(") == -1) || (strings.Index(r.Regex, ")") == -1) {
				return fmt.Errorf("gauge regex must have at least one group match for extracting the desired value from stream")
			}
		}
		r.in = make(chan string)
		logrus.Debugf("Preparing rule %s", r.Name)
		r2 = append(r2, r)
	}

	listen := fmt.Sprintf("%s:%d", opt.BindHost, opt.BindPort)
	listenPort, err := net.Listen("tcp", listen)
	if err != nil {
		panic(err)
	}

	go func() {
		logrus.Debugf("Starting Prometheus Exporter at http://%s:%d%s", opt.BindHost, opt.BindPort, opt.MetricsPath)
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
	if len(r2) == 0 {
		return fmt.Errorf("No rules defined")
	}
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
						label := ""
						c := 0.0
						if len(m) > 1 {
							c1, err1 := strconv.ParseFloat(m[1], 64)
							if err1 != nil {
								logrus.Debugf("Could not parse '%s' as float in stream", m[1])
								label = m[1]
							} else {
								c = c1
							}
						}
						if len(m) > 2 {
							c2, err2 := strconv.ParseFloat(m[2], 64)
							if err2 != nil {
								logrus.Debugf("Could not parse '%s' as float in stream", m[1])
								label = m[2]
							} else {
								c = c2
							}
						}
						if ru.Typ == TypeSummary {
							ru.metricSummary.WithLabelValues(label).Observe(c)
						} else if ru.Typ == TypeGauge {
							ru.metricGauge.WithLabelValues(label).Set(c)
						}
					}
				}
			}
		}(rul)
	}

	return nil
}
