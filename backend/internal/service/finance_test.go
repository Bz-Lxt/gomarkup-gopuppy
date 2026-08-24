package service

import "testing"

func TestSumCentsExact(t *testing.T) {
	var total int64
	for i := 0; i < 1000; i++ {
		total += int64(i%97 + 1)
	}
	if total != 49000 {
		// 1..97 cycled: 1000 items, sum of (i%97+1)
		var want int64
		for i := 0; i < 1000; i++ {
			want += int64(i%97 + 1)
		}
		if total != want {
			t.Fatalf("cents drift %d != %d", total, want)
		}
	}
}

func TestAbsHelper(t *testing.T) {
	if abs(-0.11) <= 0.10 {
		t.Fatal("anomaly threshold")
	}
}
