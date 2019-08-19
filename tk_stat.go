package templatemanager

import (
	"fmt"
	"time"
)

type QpsStat struct {
	TStart time.Time
	TStop  time.Time
	Count  int
}

func NewQpsStat(tStart time.Time, tStop time.Time, count int) *QpsStat {
	return &QpsStat{TStart: tStart, TStop: tStop, Count: count}
}

func (st *QpsStat) Duration() time.Duration {
	return st.TStop.Sub(st.TStart)
}

func (st *QpsStat) Qps() float64 {
	return float64(st.Count) / st.Duration().Seconds()
}

func (st *QpsStat) Spq() float64 {
	return 1 / st.Qps()
}


func (st *QpsStat) String() string {
	t0 := st.TStart.Format(time.RFC3339Nano)
	t1 := st.TStop.Format(time.RFC3339Nano)
	return fmt.Sprintf(`(%s, %s, d:%s), (qps:%.5f, spq:%.5f)`, t0, t1, st.Duration(), st.Qps(), st.Spq())
}

func (st *QpsStat) ShortString() string {
	return fmt.Sprintf(`%s(qps:%.3f, spq:%.3f)`, st.Duration(), st.Qps(), st.Spq())
}

