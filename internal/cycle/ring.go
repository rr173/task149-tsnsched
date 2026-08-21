package cycle

import "sort"

type Segment struct {
	Start int64
	End   int64
}

func Normalize(v, period int64) int64 {
	if period <= 0 {
		return v
	}
	v %= period
	if v < 0 {
		v += period
	}
	return v
}
func Split(start, duration, period int64) []Segment {
	if period <= 0 || duration <= 0 {
		return nil
	}
	start = Normalize(start, period)
	if duration >= period {
		return []Segment{{0, period}}
	}
	end := start + duration
	if end <= period {
		return []Segment{{start, end}}
	}
	return []Segment{{start, period}, {0, end - period}}
}
func Overlap(a, b []Segment) bool {
	for _, x := range a {
		for _, y := range b {
			if x.Start < y.End && y.Start < x.End {
				return true
			}
		}
	}
	return false
}
func Contains(segs []Segment, v int64) bool {
	for _, s := range segs {
		if v >= s.Start && v < s.End {
			return true
		}
	}
	return false
}
func Merge(segs []Segment) []Segment {
	if len(segs) < 2 {
		return segs
	}
	copySegs := append([]Segment{}, segs...)
	sort.Slice(copySegs, func(i, j int) bool { return copySegs[i].Start < copySegs[j].Start })
	out := []Segment{copySegs[0]}
	for _, s := range copySegs[1:] {
		last := &out[len(out)-1]
		if s.Start <= last.End {
			if s.End > last.End {
				last.End = s.End
			}
		} else {
			out = append(out, s)
		}
	}
	return out
}
func Width(segs []Segment) int64 {
	var n int64
	for _, s := range segs {
		n += s.End - s.Start
	}
	return n
}
