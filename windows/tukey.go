package windows

// Окно Тьюки (также известно как косинусное окно или Tukey window)
// alpha определяет долю окна с косинусоидальными переходами (0-1)
func tukeyWindow(N int, alpha float64) []float64 {
	window := make([]float64, N)

	if alpha <= 0 {
		// Прямоугольное окно
		for i := 0; i < N; i++ {
			window[i] = 1.0
		}
		return window
	}

	if alpha >= 1 {
		// Окно Хэннинга
		return hannWindow(N)
	}

	N1 := int(alpha * float64(N-1) / 2)
	N2 := N - 1 - N1

	for n := 0; n < N; n++ {
		if n < N1 {
			// Левая часть - косинусоидальный переход
			window[n] = 0.5 * (1 + math.Cos(math.Pi*(float64(n)/float64(N1)-1)))
		} else if n <= N2 {
			// Центральная часть - постоянная
			window[n] = 1.0
		} else {
			// Правая часть - косинусоидальный переход
			window[n] = 0.5 * (1 + math.Cos(math.Pi*(float64(n-N2)/float64(N1)-1)))
		}
	}
	return window
}

// ApplyTukeyWindow применяет окно Тьюки к коэффициентам фильтра
// alpha: 0 = прямоугольное окно, 1 = окно Хэннинга
func ApplyTukeyWindow(coeffs []float64, alpha float64) []float64 {
	N := len(coeffs)
	window := tukeyWindow(N, alpha)

	modifiedCoeffs := make([]float64, N)
	for i := 0; i < N; i++ {
		modifiedCoeffs[i] = coeffs[i] * window[i]
	}
	return modifiedCoeffs
}

// ApplyTukeyWindowDefault Вспомогательная функция с параметром по умолчанию для окна Тьюки
func ApplyTukeyWindowDefault(coeffs []float64) []float64 {
	// По умолчанию alpha = 0.5
	return ApplyTukeyWindow(coeffs, 0.5)
}
