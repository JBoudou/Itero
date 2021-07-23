// Itero - Online iterative vote application
// Copyright (C) 2021 Joseph Boudou
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"time"
)

type LogStat struct{}

func (self LogStat) Cmd() string {
	return "logstat"
}

func (self LogStat) String() string {
	return "Analyse Itero logs."
}

func init() {
	AddCommand(LogStat{})
}

func (self LogStat) Run(args []string) {
	timedRegex := regexp.MustCompile(`^(\S{10} \S{15}) \S `)
	fullRegex := regexp.MustCompile(`^\S{10} \S{15} H ([0-9.]{7,15}):\d+ (/\S*) .*(\d{3}) in (\S+)\s$`)
	segmentRegex := regexp.MustCompile(`^((?:/[^/]*[^/\d][^/]*){2,}?)(?:/\d+)*/[a-zA-Z0-9\-_]{9}$`)

	var firstTime time.Time
	var lastTime time.Time
	total := countTime{}
	bySt := byStatus{}
	byIp := byKeyAndStatus{}
	byRq := byKeyAndStatus{}
	var unparsedLines uint32

	rd := bufio.NewReader(os.Stdin)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			break
		}

		capture := timedRegex.FindStringSubmatch(line)
		if capture == nil {
			unparsedLines += 1
			continue
		}
		t, err := time.ParseInLocation("2006/01/02 15:04:05.000000", capture[1], time.Local)
		if err != nil {
			unparsedLines += 1
			continue
		}
		if firstTime.IsZero() {
			firstTime = t
		} else {
			lastTime = t
		}

		capture = fullRegex.FindStringSubmatch(line)
		if capture == nil {
			continue
		}
		dur, err := time.ParseDuration(capture[4])
		if err != nil {
			continue
		}
		total.Add(dur)
		status := capture[3]
		bySt.Add(status, dur)
		byIp.Add(capture[1], status, dur)

		request := capture[2]
		capture = segmentRegex.FindStringSubmatch(request)
		if capture != nil {
			request = capture[1]
		}
		byRq.Add(request, status, dur)
	}

	fmt.Printf("Server run: %v.\n", lastTime.Sub(firstTime))
	total.Display("")
	fmt.Printf("%d unparsed lines.\n", unparsedLines)

	fmt.Println("Status:")
	bySt.Display("  ")

	fmt.Println("Visitors:")
	byIp.Display("  ")
	fmt.Println("Requests:")
	byRq.Display("  ")
}

//
// countTime
//

type countTime struct {
	count    uint32
	duration time.Duration
}

func (self *countTime) Add(dur time.Duration) {
	self.count += 1
	self.duration += dur
}

func (self countTime) Display(prefix string) {
	fmt.Printf("%s%d requests in %v (%f req/sec)\n", prefix, self.count, self.duration,
		float64(self.count)/self.duration.Seconds())
}

//
// byStatus
//

type status = string

type byStatus map[status]*countTime

func (self byStatus) Add(status status, dur time.Duration) {
	ct, ok := self[status]
	if !ok {
		ct = &countTime{}
		self[status] = ct
	}
	ct.Add(dur)
}

func (self byStatus) Display(prefix string) {
	sorted := make([]status, len(self))
	i := 0
	for key := range self {
		sorted[i] = key
		i += 1
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	for _, key := range sorted {
		val := self[key]
		val.Display(prefix + key + ": ")
	}
}

//
// byKeyAndStatus
//

type byKeyAndStatus_node struct {
	total  countTime
	detail byStatus
}
type byKeyAndStatus map[string]*byKeyAndStatus_node

func (self byKeyAndStatus) Add(key string, status status, dur time.Duration) {
	bk, ok := self[key]
	if !ok {
		bk = &byKeyAndStatus_node{detail: byStatus{}}
		self[key] = bk
	}
	bk.total.Add(dur)
	bk.detail.Add(status, dur)
}

func (self byKeyAndStatus) Display(prefix string) {
	sorted := make([]string, len(self))
	i := 0
	for key := range self {
		sorted[i] = key
		i += 1
	}
	sort.Slice(sorted, func(i, j int) bool {
		return self[sorted[i]].total.duration > self[sorted[j]].total.duration
	})

	for _, key := range sorted {
		val := self[key]
		val.total.Display(prefix + key + ": ")
		val.detail.Display(prefix + "  ")
	}
}
