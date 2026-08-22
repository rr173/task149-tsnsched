package scheduler

import "testing"

// TestFrameDurationRoundsUpOnNonDivisibleRate guards the duration contract:
// when the port rate does not evenly divide the frame bit count, the slot must
// be the smallest whole-nanosecond value whose product with the rate still
// covers every bit. Truncation (the old behaviour) returned a slot one nanosecond
// too short, so the last bit straddled the reservation.
func TestFrameDurationRoundsUpOnNonDivisibleRate(t *testing.T) {
	cases := []struct {
		frame, rate, want int64
	}{
		// 100 bits / 3 bps: exact wire time = 33.(3) s, ceiling = 33333333334 ns.
		{frame: 100, rate: 3, want: 33333333334},
		// 100 bits / 7 bps: 14.2857..., ceiling = 14285714286 ns.
		{frame: 100, rate: 7, want: 14285714286},
		// 1 bit / 3 bps: 0.(3) s, ceiling = 333333334 ns.
		{frame: 1, rate: 3, want: 333333334},
		// Evenly divisible case is unchanged.
		{frame: 10, rate: 1_000_000_000, want: 10},
	}
	for _, c := range cases {
		got, err := FrameDuration(c.frame, c.rate)
		if err != nil {
			t.Fatalf("FrameDuration(%d,%d) err=%v", c.frame, c.rate, err)
		}
		if got != c.want {
			t.Fatalf("FrameDuration(%d,%d)=%d want %d", c.frame, c.rate, got, c.want)
		}
		// The defining property: the reserved slot carries the whole frame.
		if got*c.rate < c.frame*1_000_000_000 {
			t.Fatalf("FrameDuration(%d,%d)=%d under-covers: %d*%d < %d",
				c.frame, c.rate, got, got, c.rate, c.frame*1_000_000_000)
		}
	}
}

// TestFrameDurationRejectsNonPositive mirrors the input contract.
func TestFrameDurationRejectsNonPositive(t *testing.T) {
	for _, c := range [][2]int64{{0, 1}, {1, 0}, {-1, 1}, {1, -1}} {
		if _, err := FrameDuration(c[0], c[1]); err == nil {
			t.Fatalf("FrameDuration(%d,%d) expected error", c[0], c[1])
		}
	}
}
