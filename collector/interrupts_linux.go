// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !nointerrupts

package collector

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type interruptsCollector struct {
	metric *prometheus.CounterVec
}

func init() {
	Factories["interrupts"] = NewInterruptsCollector
}

// Takes a prometheus registry and returns a new Collector exposing
// interrupts stats
func NewInterruptsCollector() (Collector, error) {
	return &interruptsCollector{
		metric: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "interrupts",
				Help:      "Interrupt details.",
			},
			[]string{"CPU", "type", "info", "devices"},
		),
	}, nil
}

func (c *interruptsCollector) Update(ch chan<- prometheus.Metric) (err error) {
	interrupts, err := getInterrupts()
	if err != nil {
		return fmt.Errorf("couldn't get interrupts: %s", err)
	}
	for name, interrupt := range interrupts {
		for cpuNo, value := range interrupt.values {
			fv, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid value %s in interrupts: %s", value, err)
			}
			labels := prometheus.Labels{
				"CPU":     strconv.Itoa(cpuNo),
				"type":    name,
				"info":    interrupt.info,
				"devices": interrupt.devices,
			}
			c.metric.With(labels).Set(fv)
		}
	}
	c.metric.Collect(ch)
	return err
}

type interrupt struct {
	info    string
	devices string
	values  []string
}

func getInterrupts() (map[string]interrupt, error) {
	file, err := os.Open(procFilePath("interrupts"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return parseInterrupts(file)
}

func parseInterrupts(r io.Reader) (map[string]interrupt, error) {
	var (
		interrupts = map[string]interrupt{}
		scanner    = bufio.NewScanner(r)
	)

	if !scanner.Scan() {
		return nil, errors.New("interrupts empty")
	}
	cpuNum := len(strings.Fields(string(scanner.Text()))) // one header per cpu

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(string(line))
		if len(parts) < cpuNum+2 { // irq + one column per cpu + details,
			continue // we ignore ERR and MIS for now
		}
		intName := parts[0][:len(parts[0])-1] // remove trailing :
		intr := interrupt{
			values: parts[1:cpuNum],
		}

		if _, err := strconv.Atoi(intName); err == nil { // numeral interrupt
			intr.info = parts[cpuNum+1]
			intr.devices = strings.Join(parts[cpuNum+2:], " ")
		} else {
			intr.info = strings.Join(parts[cpuNum+1:], " ")
		}
		interrupts[intName] = intr
	}

	return interrupts, nil
}
