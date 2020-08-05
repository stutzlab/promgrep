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
	})
	rules = append(rules, MetricRule{
		Name:  "empty",
		Regex: "",
	})
	rules = append(rules, MetricRule{
		Name:  "full",
		Regex: "123abc",
	})

	err := Start(ctx, rules, PromOptions{}, test)
	assert.Nil(t, err)
	time.Sleep(1 * time.Second)

	bs := readMetrics(t)
	assert.Contains(t, bs, "promgrep_all 1")
	assert.Contains(t, bs, "promgrep_empty 1")
	assert.Contains(t, bs, "promgrep_full 1")
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
	})
	rules = append(rules, MetricRule{
		Name:  "xyz",
		Regex: "ABC.*XYZ",
	})
	rules = append(rules, MetricRule{
		Name:  "123abc",
		Regex: "123abc",
	})
	rules = append(rules, MetricRule{
		Name:  "numbers",
		Regex: "[0-9]{3,3}",
	})

	err := Start(ctx, rules, PromOptions{}, test)
	assert.Nil(t, err)
	time.Sleep(1 * time.Second)

	bs := readMetrics(t)
	assert.Contains(t, bs, "promgrep_full2 1")
	assert.Contains(t, bs, "promgrep_xyz 2")
	assert.Contains(t, bs, "promgrep_123abc 4")
	assert.Contains(t, bs, "promgrep_numbers 8")
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
	})
	rules = append(rules, MetricRule{
		Name:  "test30001",
		Regex: "TEST=([0-9]+)",
	})
	rules = append(rules, MetricRule{
		Name:  "float30123",
		Regex: "TEST=([0-9\\.]+)",
	})

	err := Start(ctx, rules, PromOptions{}, test)
	assert.Nil(t, err)
	time.Sleep(1 * time.Second)

	bs := readMetrics(t)
	assert.Contains(t, bs, "promgrep_num3 246")
	assert.Contains(t, bs, "promgrep_test30001 30001")
	assert.Contains(t, bs, "promgrep_float30123 30001.123")
}

func readMetrics(t *testing.T) string {
	resp, err := http.Get("http://localhost:8880/metrics")
	assert.Nil(t, err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body)
}
