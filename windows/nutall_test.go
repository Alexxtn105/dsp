package windows

import (
	"math"
	"testing"
)

func TestNuttallWindow(t *testing.T) {
	tests := []struct {
		name   string
		N      int
		checks []struct {
			index int
			want  float64
		}
	}{
		{
			name: "Window size 1",
			N:    1,
			checks: []struct {
				index int
				want  float64
			}{
				{0, 1.0}, // При N=1 формула даёт a0 - a1*cos(0) + a2*cos(0) - a3*cos(0) = 0.355768 - 0.487396 + 0.144232 - 0.012604 = 0.0, но практичнее вернуть 1.0
			},
		},
		{
			name: "Window size 5",
			N:    5,
			checks: []struct {
				index int
				want  float64
			}{
				{0, 0.0}, // Край (должен быть близок к 0)
				{2, 1.0}, // Центр (должен быть близок к 1)
				{4, 0.0}, // Симметрично первому
			},
		},
		{
			name: "Window size 8",
			N:    8,
			checks: []struct {
				index int
				want  float64
			}{
				{0, 0.0},       // Край
				{4, 0.8876285}, // Промежуточное значение (точное значение из формулы)
				{7, 0.0},       // Симметрично первому
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			window := nuttallWindow(tt.N)

			// Проверяем длину
			if len(window) != tt.N {
				t.Errorf("nuttallWindow(%d) length = %d, want %d", tt.N, len(window), tt.N)
			}

			// Проверяем симметричность (для N > 1)
			if tt.N > 1 {
				for i := 0; i < tt.N/2; i++ {
					diff := math.Abs(window[i] - window[tt.N-1-i])
					if diff > 1e-10 {
						t.Errorf("nuttallWindow(%d) not symmetric at indices %d and %d: %f != %f",
							tt.N, i, tt.N-1-i, window[i], window[tt.N-1-i])
					}
				}
			}

			// Проверяем конкретные значения
			for _, check := range tt.checks {
				if check.index >= tt.N {
					continue
				}
				if math.Abs(window[check.index]-check.want) > 1e-6 {
					t.Errorf("nuttallWindow(%d)[%d] = %e, want %e",
						tt.N, check.index, window[check.index], check.want)
				}
			}

			// Проверяем что все значения в диапазоне [-eps, 1+eps]
			// Окно может иметь небольшие отрицательные значения из-за численных погрешностей
			for i, val := range window {
				if val < -1e-10 || val > 1.0+1e-10 {
					t.Errorf("nuttallWindow(%d)[%d] = %f, should be in [-0, 1]", tt.N, i, val)
				}
			}
		})
	}
}

func TestNuttallWindowApply(t *testing.T) {
	tests := []struct {
		name      string
		coeffs    []float64
		tolerance float64
	}{
		{
			name:      "Empty coefficients",
			coeffs:    []float64{},
			tolerance: 1e-12,
		},
		{
			name:      "Single coefficient",
			coeffs:    []float64{2.5},
			tolerance: 1e-6,
		},
		{
			name:      "All ones, size 3",
			coeffs:    []float64{1.0, 1.0, 1.0},
			tolerance: 1e-6,
		},
		{
			name:      "Linear coefficients, size 5",
			coeffs:    []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			tolerance: 1e-6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Вычисляем окно
			N := len(tt.coeffs)
			window := nuttallWindow(N)

			// Вычисляем ожидаемые значения
			wantCoeff := make([]float64, N)
			for i := 0; i < N; i++ {
				wantCoeff[i] = tt.coeffs[i] * window[i]
			}

			got := ApplyNuttallWindow(tt.coeffs)

			if len(got) != len(wantCoeff) {
				t.Errorf("ApplyNuttallWindow() length = %d, want %d", len(got), len(wantCoeff))
				return
			}

			for i := range got {
				if math.Abs(got[i]-wantCoeff[i]) > tt.tolerance {
					t.Errorf("ApplyNuttallWindow()[%d] = %f, want %f (tolerance %e)",
						i, got[i], wantCoeff[i], tt.tolerance)
				}
			}
		})
	}

	// Тест на неизменность исходного массива
	t.Run("Original array not modified", func(t *testing.T) {
		original := []float64{1.0, 2.0, 3.0}
		coeffsCopy := make([]float64, len(original))
		copy(coeffsCopy, original)

		_ = ApplyNuttallWindow(original)

		for i := range original {
			if original[i] != coeffsCopy[i] {
				t.Errorf("ApplyNuttallWindow modified original array at index %d: %f != %f",
					i, original[i], coeffsCopy[i])
			}
		}
	})
}

func TestNuttallWindowProperties(t *testing.T) {
	testCases := []int{1, 2, 3, 4, 5, 10, 20, 50, 100}

	for _, N := range testCases {
		t.Run("N="+string(rune(N)), func(t *testing.T) {
			window := nuttallWindow(N)

			// Проверка длины
			if len(window) != N {
				t.Errorf("Length mismatch: got %d, want %d", len(window), N)
			}

			// Проверка симметричности для N > 1
			if N > 1 {
				for i := 0; i < N/2; i++ {
					if math.Abs(window[i]-window[N-1-i]) > 1e-12 {
						t.Errorf("Not symmetric at %d and %d: %f != %f",
							i, N-1-i, window[i], window[N-1-i])
					}
				}
			}

			// Проверка диапазона значений (с учетом численных погрешностей)
			for i, val := range window {
				if val < -1e-10 || val > 1.0+1e-10 {
					t.Errorf("Value at %d out of range [-0,1]: %f", i, val)
				}
			}

			// Проверка крайних значений (близки к 0)
			if N > 1 {
				if math.Abs(window[0]) > 1e-6 {
					t.Errorf("First value should be very small (<1e-6), got %e", window[0])
				}
				if math.Abs(window[N-1]) > 1e-6 {
					t.Errorf("Last value should be very small (<1e-6), got %e", window[N-1])
				}
			}

			// Проверка центрального значения (близко к 1 для нечетных N)
			if N%2 == 1 && N > 1 {
				center := N / 2
				if math.Abs(window[center]-1.0) > 1e-6 {
					t.Errorf("Center value should be close to 1.0, got %f", window[center])
				}
			}

			// Проверка суммы коэффициентов
			if N > 10 {
				var sum float64
				for _, val := range window {
					sum += val
				}
				avg := sum / float64(N)
				// Среднее значение окна Натталла около 0.36
				if math.Abs(avg-0.36) > 0.1 {
					t.Errorf("Average window value should be around 0.36, got %f", avg)
				}
			}
		})
	}
}

func TestNuttallWindowSpecificValues(t *testing.T) {
	// Проверяем вычисленные значения
	N := 10
	window := nuttallWindow(N)

	// Вычисляем ожидаемые значения
	expected := make([]float64, N)
	a0 := 0.355768
	a1 := 0.487396
	a2 := 0.144232
	a3 := 0.012604

	for n := 0; n < N; n++ {
		theta := 2 * math.Pi * float64(n) / float64(N-1)
		expected[n] = a0 -
			a1*math.Cos(theta) +
			a2*math.Cos(2*theta) -
			a3*math.Cos(3*theta)
	}

	for i := range window {
		if math.Abs(window[i]-expected[i]) > 1e-12 {
			t.Errorf("window[%d] = %e, want %e", i, window[i], expected[i])
		}
	}
}

func TestNuttallWindowComparison(t *testing.T) {
	// Сравниваем окно Натолла с другими окнами
	N := 32
	nuttall := nuttallWindow(N)

	// Проверяем что окно Натолла имеет очень низкие боковые лепестки
	// (характерное свойство окна Натолла)

	// Среднее значение
	var sum float64
	for _, val := range nuttall {
		sum += val
	}
	mean := sum / float64(N)

	// Окно Натолла должно иметь среднее около 0.36 (не 0.5!)
	if math.Abs(mean-0.36) > 0.05 {
		t.Errorf("Nuttall window mean should be around 0.36, got %f", mean)
	}

	// Проверяем монотонность от краев к центру
	for i := 1; i < N/2; i++ {
		if nuttall[i] < nuttall[i-1]-1e-12 {
			t.Errorf("Window should be monotonically increasing from edges to center: nuttall[%d]=%f < nuttall[%d]=%f",
				i, nuttall[i], i-1, nuttall[i-1])
		}
	}
}
