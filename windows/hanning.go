package windows

// Окно Хэннинга (также известно как Hann window)
func hannWindow(N int) []float64 {
	window := make([]float64, N)
	alpha := 0.5

	for n := 0; n < N; n++ {
		x := math.Pi * 2 * float64(n) / float64(N-1)
		window[n] = alpha - alpha*math.Cos(x)
	}
	return window
}

// ApplyHannWindow применяет окно Хэннинга к коэффициентам фильтра
func ApplyHannWindow(coeffs []float64) []float64 {
	N := len(coeffs)
	window := hannWindow(N)

	modifiedCoeffs := make([]float64, N)
	for i := 0; i < N; i++ {
		modifiedCoeffs[i] = coeffs[i] * window[i]
	}
	return modifiedCoeffs
}
