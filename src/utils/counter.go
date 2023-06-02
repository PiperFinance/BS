package utils

import (
	"fmt"
	"sync"
	"time"
)

const (
	MAX_RESULTS = 200
)

type TimeFrameCounter struct {
	Name        string
	Window      time.Duration
	EndsAt      time.Time
	StartedAt   time.Time
	Count       uint64
	LastResults []uint64
}

func (tfc *TimeFrameCounter) NewCall(t time.Time) {
	if t.Local().After(tfc.EndsAt) {
		// Window is finished !
		tfc.LastResults = append(tfc.LastResults, tfc.Count)
		tfc.EndsAt = t.Add(tfc.Window)
		tfc.StartedAt = t
		if len(tfc.LastResults) > MAX_RESULTS {
			tfc.LastResults = tfc.LastResults[1:len(tfc.LastResults)]
		}
		tfc.Count = 0
	}
	tfc.Count++
}

type DebugCounter struct {
	LastCallTime    map[int64]time.Time
	TimeFrames      map[int64][]TimeFrameCounter
	timeFramesCount map[int64]int
	mutex           sync.Mutex
}

func NewDebugCounter(chains []int64, timeFrames ...time.Duration) *DebugCounter {
	t := time.Now()
	r := new(DebugCounter)
	r.mutex = sync.Mutex{}
	r.timeFramesCount = make(map[int64]int, len(chains))
	r.LastCallTime = make(map[int64]time.Time, len(chains))
	r.TimeFrames = make(map[int64][]TimeFrameCounter, len(chains))
	for _, chain := range chains {

		tfs := make([]TimeFrameCounter, len(timeFrames))
		for i, tf := range timeFrames {
			tfs[i] = TimeFrameCounter{
				Name:      tf.String(),
				Window:    tf,
				StartedAt: t,
				EndsAt:    t.Add(tf),
			}
		}
		r.TimeFrames[chain] = tfs
		r.timeFramesCount[chain] = len(tfs)
		r.LastCallTime[chain] = t
	}
	return r
}

func (cc *DebugCounter) Add(chain int64, count ...int) {
	incrCount := 1
	if len(count) > 1 {
		incrCount = count[0]
	}
	i := 0
	t := time.Now()
	cc.mutex.Lock()
	cc.LastCallTime[chain] = t
	for i < cc.timeFramesCount[chain] {
		cc.TimeFrames[chain][i].NewCall(t)
		i += incrCount
	}
	cc.mutex.Unlock()
}

func (cc *DebugCounter) StatusChain(chain int64) string {
	_x := cc.TimeFrames[chain]
	x := _x[len(_x)-1]
	rate := float64(x.Count) / x.Window.Seconds()
	return fmt.Sprintf("%f", rate)
}

func (cc *DebugCounter) StatusChainIndexOf(index int, chain int64) string {
	x := cc.TimeFrames[chain][index]
	rate := float64(x.Count) / x.Window.Seconds()
	return fmt.Sprintf("%f", rate)
}
