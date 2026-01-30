package windows

import "math"

// Окно Натолла (4-х членное)
func nuttallWindow(N int) []float64 {
	window := make([]float64, N)

	// Коэффициенты для 4-х членного окна Натолла
	a0 := 0.355768
	a1 := 0.487396
	a2 := 0.144232
	a3 := 0.012604

	for n := 0; n < N; n++ {
		theta := 2 * math.Pi * float64(n) / float64(N-1)
		window[n] = a0 -
			a1*math.Cos(theta) +
			a2*math.Cos(2*theta) -
			a3*math.Cos(3*theta)
	}
	return window
}

// ApplyNuttallWindow применяет окно Натолла к коэффициентам фильтра
func ApplyNuttallWindow(coeffs []float64) []float64 {
	N := len(coeffs)
	window := nuttallWindow(N)

	modifiedCoeffs := make([]float64, N)
	for i := 0; i < N; i++ {
		modifiedCoeffs[i] = coeffs[i] * window[i]
	}
	return modifiedCoeffs
}
