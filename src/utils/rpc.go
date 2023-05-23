package utils

import "time"

const (
	MAX_RESULTS = 100
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

type CallCounter struct {
	LastCallTime    time.Time
	TimeFrames      []TimeFrameCounter
	timeFramesCount int
}

func NewCallCounter(timeFrames ...time.Duration) *CallCounter {
	t := time.Now()
	r := new(CallCounter)
	tfs := make([]TimeFrameCounter, len(timeFrames))
	for i, tf := range timeFrames {
		tfs[i] = TimeFrameCounter{
			Name:      tf.String(),
			Window:    tf,
			StartedAt: t,
			EndsAt:    t.Add(tf),
		}
	}
	r.TimeFrames = tfs
	r.timeFramesCount = len(tfs)
	return r
}

func (cc *CallCounter) Add() {
	i := 0
	t := time.Now()
	cc.LastCallTime = t
	for i < cc.timeFramesCount {
		i++
		cc.TimeFrames[i].NewCall(t)
	}
}

func (cc *CallCounter) Status() {
	// TODO
	// return cc
}
