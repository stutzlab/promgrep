package main

import (
	"flag"
)

type option struct {
	bindHost    string
	bindPort    uint
	metricsPath string

	countMetricRules arrayFlags
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {

	opt := option{}

	flag.UintVar(&opt.bindPort, "port", 8880, "Prometheus exporter port. defaults to 8880")
	flag.StringVar(&opt.bindHost, "host", "0.0.0.0", "Prometheus exporter bind host. defaults to 0.0.0.0")
	flag.StringVar(&opt.metricsPath, "path", "/metrics", "Prometheus exporter port. defaults to /metric")

	flag.Var(&opt.countMetricRules, "count", "Count metrics regex grep config. Ex.: 'question@question\\sfinished' for a counter under metric 'question' that gets incremented when 'question finished' substring is found in stdin stream. See more info at http://github.com/stutzlab/promgrep")

}
