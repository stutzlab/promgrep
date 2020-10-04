package main

import (
	"context"
	"flag"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/stutzlab/promgrep"
)

type option struct {
	bindHost    string
	bindPort    uint
	metricsPath string
	outputMode  string

	summaryMetricRules arrayFlags
	gaugeMetricRules   arrayFlags
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
	logrus.SetLevel(logrus.InfoLevel)

	opt := option{}

	flag.UintVar(&opt.bindPort, "port", 8880, "Prometheus exporter port. defaults to 8880")
	flag.StringVar(&opt.bindHost, "host", "0.0.0.0", "Prometheus exporter bind host. defaults to 0.0.0.0")
	flag.StringVar(&opt.metricsPath, "path", "/metrics", "Prometheus exporter port. defaults to /metric")
	flag.StringVar(&opt.outputMode, "output", "match", "Defines what to output to stdout. One of 'none' (silent output), 'match' (will print the matched lines) or 'all' (will passthrough all input to output stream). Defaults to 'match'")

	flag.Var(&opt.summaryMetricRules, "summary", "Summarized count metrics regex grep config. Ex.: 'question@question\\sfinished\\s([0-9]+)ms' will expose a metric with a counter of number of questions finished and another with the sum of question latencies over time. See more info at http://github.com/stutzlab/promgrep")
	flag.Var(&opt.gaugeMetricRules, "gauge", "Gauge metrics regex grep config. Ex.: 'temperature@temp-now-is-([0-9\\.]+)degrees' will expose a metric with the temperature gotten from the stream. See more info at http://github.com/stutzlab/promgrep")
	flag.Parse()

	rules := make([]promgrep.MetricRule, 0)

	for _, mr := range opt.summaryMetricRules {
		rules = addRule(mr, promgrep.TypeSummary, rules)
	}

	for _, mr := range opt.gaugeMetricRules {
		rules = addRule(mr, promgrep.TypeGauge, rules)
	}

	ctx := context.Background()

	opt2 := promgrep.PromOptions{
		BindHost:    opt.bindHost,
		BindPort:    opt.bindPort,
		MetricsPath: opt.metricsPath,
		Output:      promgrep.Output(opt.outputMode),
	}

	err := promgrep.Start(ctx, rules, opt2, os.Stdin, os.Stdout)
	if err != nil {
		logrus.Warnf("Could not start promgrep: %s", err)
		os.Exit(1)
	}
	<-ctx.Done()
}

func addRule(ruleStr string, typ promgrep.MetricType, rules []promgrep.MetricRule) []promgrep.MetricRule {
	i := strings.Index(ruleStr, "@")
	if i == -1 {
		logrus.Warnf("Metric definition %s must be in format [name]@[regex]", ruleStr)
		os.Exit(1)
	}
	name := ruleStr[0:i]
	regex := ruleStr[i+1:]
	rules = append(rules, promgrep.MetricRule{
		Name:  name,
		Regex: regex,
		Typ:   typ,
	})

	return rules
}
