package windows

import (
	"math"
)

// Функция Бесселя первого рода нулевого порядка
func besselI0(x float64) float64 {
	if x == 0 {
		return 1.0
	}

	// Аппроксимация с использованием ряда Тейлора
	var result float64 = 1.0
	var term float64 = 1.0
	xSquaredOver4 := x * x / 4.0

	for k := 1; k <= 20; k++ { // 20 итераций достаточно для хорошей точности
		term *= xSquaredOver4 / float64(k*k)
		result += term
	}

	return result
}

// Окно Кайзера
func kaiserWindow(N int, beta float64) []float64 {
	if N <= 0 {
		return []float64{}
	}

	window := make([]float64, N)

	// Нормализующий множитель
	denominator := besselI0(beta)

	// Для N=1 особый случай
	if N == 1 {
		window[0] = 1.0
		return window
	}

	// Вычисляем значения окна
	M := float64(N - 1)
	for n := 0; n < N; n++ {
		x := 2*float64(n)/M - 1 // x ∈ [-1, 1]
		numerator := beta * math.Sqrt(1-x*x)
		window[n] = besselI0(numerator) / denominator
	}

	return window
}

// ApplyKaiserWindow применяет окно Кайзера к коэффициентам фильтра
func ApplyKaiserWindow(coeffs []float64, beta float64) []float64 {
	N := len(coeffs)
	if N == 0 {
		return []float64{}
	}

	window := kaiserWindow(N, beta)
	modifiedCoeffs := make([]float64, N)

	for i := 0; i < N; i++ {
		modifiedCoeffs[i] = coeffs[i] * window[i]
	}

	return modifiedCoeffs
}
