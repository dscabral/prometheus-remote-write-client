package prometheus

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type LabelList []Label
type HeaderList []header
type Dp Datapoint

type header struct {
	name  string
	value string
}

func (t *LabelList) String() string {
	var labels [][]string
	for _, v := range []Label(*t) {
		labels = append(labels, []string{v.Name, v.Value})
	}
	return fmt.Sprintf("%v", labels)
}

func (t *LabelList) Set(value string) error {
	labelPair := strings.Split(value, ":")
	if len(labelPair) != 2 {
		return fmt.Errorf("incorrect number of arguments to '-t': %d", len(labelPair))
	}

	label := Label{
		Name:  labelPair[0],
		Value: labelPair[1],
	}

	*t = append(*t, label)

	return nil
}

func (h *HeaderList) String() string {
	var headers [][]string
	for _, v := range []header(*h) {
		headers = append(headers, []string{v.name, v.value})
	}
	return fmt.Sprintf("%v", headers)
}

func (h *HeaderList) Set(value string) error {
	firstSplit := strings.Index(value, ":")
	if firstSplit == -1 {
		return fmt.Errorf("header missing separating colon: '%v'", value)
	}

	*h = append(*h, header{
		name:  strings.TrimSpace(value[:firstSplit]),
		value: strings.TrimSpace(value[firstSplit+1:]),
	})

	return nil
}

func (d *Dp) String() string {
	return fmt.Sprintf("%v", []string{d.Timestamp.String(), fmt.Sprintf("%v", d.Value)})
}

func (d *Dp) Set(value string) error {
	dp := strings.Split(value, ",")
	if len(dp) != 2 {
		return fmt.Errorf("incorrect number of arguments to '-d': %d", len(dp))
	}

	var ts time.Time
	if strings.ToLower(dp[0]) == "now" {
		ts = time.Now()
	} else {
		i, err := strconv.Atoi(dp[0])
		if err != nil {
			return fmt.Errorf("unable to parse timestamp: %s", dp[1])
		}
		ts = time.Unix(int64(i), 0)
	}

	val, err := strconv.ParseFloat(dp[1], 64)
	if err != nil {
		return fmt.Errorf("unable to parse value as float64: %s", dp[0])
	}

	d.Timestamp = ts
	d.Value = val

	return nil
}

