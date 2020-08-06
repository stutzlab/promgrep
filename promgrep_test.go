package promgrep

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	// logrus.SetLevel(logrus.DebugLevel)

	test := strings.NewReader("abc123abc123\n")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rules := make([]MetricRule, 0)
	rules = append(rules, MetricRule{
		Name:  "all",
		Regex: ".*",
		Typ:   TypeSummary,
	})
	rules = append(rules, MetricRule{
		Name:  "empty",
		Regex: "",
		Typ:   TypeSummary,
	})
	rules = append(rules, MetricRule{
		Name:  "full",
		Regex: "123abc",
		Typ:   TypeSummary,
	})

	err := Start(ctx, rules, PromOptions{}, test)
	assert.Nil(t, err)
	time.Sleep(1 * time.Second)

	bs := readMetrics(t)
	assert.Contains(t, bs, "promgrep_all_sum{label=\"\"} 0")
	assert.Contains(t, bs, "promgrep_all_count{label=\"\"} 1")
	assert.Contains(t, bs, "promgrep_empty_sum{label=\"\"} 0")
	assert.Contains(t, bs, "promgrep_empty_count{label=\"\"} 1")
	assert.Contains(t, bs, "promgrep_full_sum{label=\"\"} 0")
	assert.Contains(t, bs, "promgrep_full_count{label=\"\"} 1")
}

func TestSimpleCounter(t *testing.T) {
	// logrus.SetLevel(logrus.DebugLevel)

	test := strings.NewReader("abc123abc123ABCasfasfads123abcasdfadfsa123abcsdfasdfas123abcXYZaskdfjakljdhf\naslkdfj asldkjfh lksjdhf alkjdf lkasdfhABCalsdfjha678sldkjf432hsadlk098fXYY ajkldhfjlashdfljkadshABC lshfalksjfhaklsdhf XYZ")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rules := make([]MetricRule, 0)
	rules = append(rules, MetricRule{
		Name:  "full2",
		Regex: "sfasfads123abcasdfadfsa1",
		Typ:   TypeSummary,
	})
	rules = append(rules, MetricRule{
		Name:  "xyz",
		Regex: "ABC.*XYZ",
		Typ:   TypeSummary,
	})
	rules = append(rules, MetricRule{
		Name:  "123abc",
		Regex: "123abc",
		Typ:   TypeSummary,
	})
	rules = append(rules, MetricRule{
		Name:  "numbers",
		Regex: "[0-9]{3,3}",
		Typ:   TypeSummary,
	})
	rules = append(rules, MetricRule{
		Name:  "numbers_summed",
		Regex: "([0-9]{3,3})",
		Typ:   TypeSummary,
	})

	err := Start(ctx, rules, PromOptions{}, test)
	assert.Nil(t, err)
	time.Sleep(1 * time.Second)

	bs := readMetrics(t)
	assert.Contains(t, bs, "promgrep_full2_count{label=\"\"} 1")
	assert.Contains(t, bs, "promgrep_full2_sum{label=\"\"} 0")
	assert.Contains(t, bs, "promgrep_xyz_count{label=\"\"} 2")
	assert.Contains(t, bs, "promgrep_xyz_sum{label=\"\"} 0")
	assert.Contains(t, bs, "promgrep_123abc_count{label=\"\"} 4")
	assert.Contains(t, bs, "promgrep_123abc_sum{label=\"\"} 0")
	assert.Contains(t, bs, "promgrep_numbers_count{label=\"\"} 8")
	assert.Contains(t, bs, "promgrep_numbers_sum{label=\"\"} 0")
	assert.Contains(t, bs, "promgrep_numbers_summed_count{label=\"\"} 8")
	assert.Contains(t, bs, "promgrep_numbers_summed_sum{label=\"\"} 1823")
}

func TestCounterExtract(t *testing.T) {
	// logrus.SetLevel(logrus.DebugLevel)

	test := strings.NewReader("abc123abc123ABCasfasfads123abcasdTEST=10000fadfsa123abcsdfasdfas123abcXYZaskdfjakTEST=20000ljdhf\naslkdfj asldkjfh lksjdhf alkjdf lkasdfTEST=1.123hABCalsdfjha=678sldkjf432hsadlk098fXYY ajkldhfjlashdfljkadshABC lshfalksjfhaklsdhf XYZ")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rules := make([]MetricRule, 0)
	rules = append(rules, MetricRule{
		Name:  "num3",
		Regex: "abc([0-9]+)",
		Typ:   TypeSummary,
	})
	rules = append(rules, MetricRule{
		Name:  "test30001",
		Regex: "TEST=([0-9]+)",
		Typ:   TypeSummary,
	})
	rules = append(rules, MetricRule{
		Name:  "float98",
		Regex: "TEST=([0-9\\.]+)",
		Typ:   TypeGauge,
	})

	err := Start(ctx, rules, PromOptions{}, test)
	assert.Nil(t, err)
	time.Sleep(1 * time.Second)

	bs := readMetrics(t)
	assert.Contains(t, bs, "promgrep_num3_count{label=\"\"} 2")
	assert.Contains(t, bs, "promgrep_num3_sum{label=\"\"} 246")
	assert.Contains(t, bs, "promgrep_test30001_count{label=\"\"} 3")
	assert.Contains(t, bs, "promgrep_test30001_sum{label=\"\"} 30001")
	assert.Contains(t, bs, "promgrep_float98{label=\"\"} 1.123")
}

func TestCounterExtractLabel(t *testing.T) {
	// logrus.SetLevel(logrus.DebugLevel)

	test := strings.NewReader("abc123abc123ABCasfasfads123abcasdTEST=10000fadfsa123abcsdfasdfas123abc XYZ=ABRACADABRA fjakTEST=20000ljdhf\naslkdfj asldkjfh lksjdhf alkjdf lkasdfTEST=1.123hABCalsdfjha=678sldkjf432hsadlk098fXYY ajkldhfjlashdfljkadshABC lshfalksjfhaklsdhf XYZ")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rules := make([]MetricRule, 0)
	rules = append(rules, MetricRule{
		Name:  "sum4a",
		Regex: "abc([0-9]+).*XYZ=([a-zA-Z]+)",
		Typ:   TypeSummary,
	})
	rules = append(rules, MetricRule{
		Name:  "sum4b",
		Regex: "XYZ=([a-zA-Z]+).*TEST=([0-9]+)",
		Typ:   TypeSummary,
	})
	rules = append(rules, MetricRule{
		Name:  "gauge4c",
		Regex: "abc([0-9]+).*XYZ=([a-zA-Z]+)",
		Typ:   TypeGauge,
	})
	rules = append(rules, MetricRule{
		Name:  "gauge4d",
		Regex: "XYZ=([a-zA-Z]+).*TEST=([0-9]+)",
		Typ:   TypeGauge,
	})

	err := Start(ctx, rules, PromOptions{}, test)
	assert.Nil(t, err)
	time.Sleep(1 * time.Second)

	bs := readMetrics(t)
	assert.Contains(t, bs, "promgrep_sum4a_sum{label=\"ABRACADABRA\"} 123")
	assert.Contains(t, bs, "promgrep_sum4a_count{label=\"ABRACADABRA\"} 1")
	assert.Contains(t, bs, "promgrep_sum4b_sum{label=\"ABRACADABRA\"} 20000")
	assert.Contains(t, bs, "promgrep_sum4b_count{label=\"ABRACADABRA\"} 1")
	assert.Contains(t, bs, "promgrep_gauge4c{label=\"ABRACADABRA\"} 123")
	assert.Contains(t, bs, "promgrep_gauge4d{label=\"ABRACADABRA\"} 20000")
}

func readMetrics(t *testing.T) string {
	resp, err := http.Get("http://localhost:8880/metrics")
	assert.Nil(t, err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body)
}
