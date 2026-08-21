package cycle

type Guard struct {
	Before int64
	After  int64
}

func Expand(start, duration, period int64, g Guard) []Segment {
	return Split(start-g.Before, duration+g.Before+g.After, period)
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
