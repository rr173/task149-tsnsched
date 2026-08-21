package cycle

type Guard struct {
	Before int64
	After  int64
}

func Expand(start, duration, period int64, g Guard) []Segment {
	return Split(start-g.Before, duration+g.Before+g.After, period)
}

// Reserved returns the ring segments a transmission occupies once its guard
// bands are folded in: the [before] guard in front of the slot, the slot
// itself, and the [after] guard trailing it, all wrapped onto the cycle.
// This is the single source of truth for the "reserved range" used by
// conflict detection, capacity accounting and windowing so that the guard
// semantics are identical across all three stages.
func Reserved(start, duration, period int64, g Guard) []Segment {
	if duration < 0 {
		duration = 0
	}
	if g.Before < 0 {
		g.Before = 0
	}
	if g.After < 0 {
		g.After = 0
	}
	return Expand(start, duration, period, g)
}

func Fits(start, duration, period int64, g Guard) bool {
	return Width(Expand(start, duration, period, g)) < period
}

func Distance(a, b int64, period int64) int64 {
	a = Normalize(a, period)
	b = Normalize(b, period)
	d := b - a
	if d < 0 {
		d += period
	}
	return d
}

func ClampWindow(from, to, period int64) (int64, int64) {
	if period <= 0 {
		return from, to
	}
	return Normalize(from, period), Normalize(to, period)
}
