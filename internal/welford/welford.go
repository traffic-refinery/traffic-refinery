// Simple package used to calculate a average and standard deviation of a series of
// values online, using the welford method
package welford

import (
	"math"
)


// Struct used to calculate welford's method
type Welford struct {
	N      float64
	Avg    float64
	m2     float64
	StdDev float64
	Var    float64
}


// Reset resets the welford computation to the initial state (i.e. all zeros)
func (wf *Welford) Reset() {
	wf.N = 0
	wf.Avg = 0
	wf.m2 = 0
	wf.StdDev = 0
	wf.Var = 0
}


// CheckAndAddValue adds a value to the welford online computation
func (wf *Welford) AddValue(val float64) {
	if wf.N == 0 {
		wf.N = 1
		wf.Avg = val
	} else {
		wf.N += 1
		delta := val - wf.Avg
		wf.Avg = wf.Avg + delta/(wf.N)
		delta2 := val - wf.Avg
		wf.m2 = wf.m2 + delta*delta2
		wf.Var = wf.m2 / (wf.N - 1)
		wf.StdDev = math.Sqrt(wf.m2)
	}
}


// CheckAndAddValue adds a value to the welford online computation only if
// val is less than maxStdDev times the standard deviation and less than maxIat.
// Returns True if the value is added. False otherwise
func (wf *Welford) CheckAndAddValue(val, maxStdDev, maxVal float64) bool {
	if wf.N == 0 {
		wf.N = 1
		wf.Avg = val
	} else {
		delta := val - wf.Avg
		mean := wf.Avg + delta/(wf.N+1)
		delta2 := val - mean
		m2 := wf.m2 + delta*delta2
		stddev := math.Sqrt(m2 / wf.N)
		if val > maxStdDev*stddev && val > maxVal {
			return false
		}
		wf.Var = m2 / wf.N
		wf.N += 1
		wf.Avg = mean
		wf.StdDev = stddev
	}
	return true
}
