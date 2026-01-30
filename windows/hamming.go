package windows

import "math"

// Окно Хэмминга
func hammingWindow(N int) []float64 {
	window := make([]float64, N)
	alpha := 0.54
	beta := 0.46

	for n := 0; n < N; n++ {
		x := math.Pi * 2 * float64(n) / float64(N-1)
		window[n] = alpha - beta*math.Cos(x)
	}
	return window
}

// ApplyHammingWindow применяет окно Хэмминга к коэффициентам фильтра
func ApplyHammingWindow(coeffs []float64) []float64 {
	N := len(coeffs)
	window := hammingWindow(N)

	modifiedCoeffs := make([]float64, N)
	for i := 0; i < N; i++ {
		modifiedCoeffs[i] = coeffs[i] * window[i]
	}
	return modifiedCoeffs
}
