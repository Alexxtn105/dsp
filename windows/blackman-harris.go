package windows

import "math"

// blackmanHarrisWindow генерирует коэффициенты окна Блэкмана-Харриса
func blackmanHarrisWindow(N int) []float64 {
	window := make([]float64, N)
	a0, a1, a2, a3 := 0.35875, 0.48829, 0.14128, 0.01168

	for n := 0; n < N; n++ {
		x := math.Pi * 2 * float64(n) / float64(N-1)
		window[n] = a0 -
			a1*math.Cos(x) +
			a2*math.Cos(2*x) -
			a3*math.Cos(3*x)
	}
	return window
}

// ApplyBlackmanHarrisWindow применяется к исходным коэффициентам фильтра
func ApplyBlackmanHarrisWindow(coeffs []float64) []float64 {
	N := len(coeffs)
	window := blackmanHarrisWindow(N)

	modifiedCoeffs := make([]float64, N)
	for i := 0; i < N; i++ {
		modifiedCoeffs[i] = coeffs[i] * window[i]
	}
	return modifiedCoeffs
}
