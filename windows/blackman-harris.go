package windows

import "math"

// blackmanHarrisWindow генерирует коэффициенты окна Блэкмана-Харриса
func blackmanHarrisWindow(N int) []float64 {
	if N <= 0 {
		return []float64{}
	}
	if N == 1 {
		return []float64{1.0}
	}

	window := make([]float64, N)
	// Коэффициенты для минимальной 4-членной версии окна Блэкмана-Харриса
	a0 := 0.35875
	a1 := 0.48829
	a2 := 0.14128
	a3 := 0.01168

	for n := 0; n < N; n++ {
		angle := 2.0 * math.Pi * float64(n) / float64(N-1)
		window[n] = a0 -
			a1*math.Cos(angle) +
			a2*math.Cos(2*angle) -
			a3*math.Cos(3*angle)
	}
	return window
}

// ApplyBlackmanHarrisWindow применяется к исходным коэффициентам фильтра
func ApplyBlackmanHarrisWindow(coeffs []float64) []float64 {
	N := len(coeffs)
	if N == 0 {
		return []float64{}
	}
	if N == 1 {
		// Для одного элемента возвращаем как есть (окно = [1.0])
		return []float64{coeffs[0]}
	}

	window := blackmanHarrisWindow(N)
	modifiedCoeffs := make([]float64, N)
	for i := 0; i < N; i++ {
		modifiedCoeffs[i] = coeffs[i] * window[i]
	}
	return modifiedCoeffs
}
