package windows

import (
	"math"
	"testing"
)

func TestBesselI0(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
		tol      float64
	}{
		{
			name:     "Zero",
			input:    0.0,
			expected: 1.0,
			tol:      1e-12,
		},
		{
			name:     "Small positive",
			input:    0.5,
			expected: 1.0634833707413236,
			tol:      1e-12,
		},
		{
			name:     "Medium positive",
			input:    2.0,
			expected: 2.2795853023360668,
			tol:      1e-12,
		},
		{
			name:     "Large positive",
			input:    10.0,
			expected: 2815.716628466254,
			tol:      1e-8,
		},
		{
			name:     "Negative (функция четная)",
			input:    -3.0,
			expected: 4.880792585865024,
			tol:      1e-12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := besselI0(tt.input)
			if math.Abs(got-tt.expected) > tt.tol {
				t.Errorf("besselI0(%f) = %e, want %e (tolerance %e)",
					tt.input, got, tt.expected, tt.tol)
			}
		})
	}

	// Проверка симметричности (четность функции)
	t.Run("Symmetry check", func(t *testing.T) {
		testValues := []float64{0.1, 0.5, 1.0, 2.0, 5.0}
		for _, x := range testValues {
			pos := besselI0(x)
			neg := besselI0(-x)
			if math.Abs(pos-neg) > 1e-12 {
				t.Errorf("besselI0 not symmetric at x=%f: I0(%f)=%e, I0(%f)=%e",
					x, x, pos, -x, neg)
			}
		}
	})
}

func TestKaiserWindow(t *testing.T) {
	tests := []struct {
		name   string
		N      int
		beta   float64
		checks []struct {
			index int
			want  float64
		}
	}{
		{
			name: "Window size 1, beta=0",
			N:    1,
			beta: 0.0,
			checks: []struct {
				index int
				want  float64
			}{
				{0, 1.0},
			},
		},
		{
			name: "Window size 5, beta=0 (прямоугольное окно)",
			N:    5,
			beta: 0.0,
			checks: []struct {
				index int
				want  float64
			}{
				{0, 1.0},
				{1, 1.0},
				{2, 1.0},
				{3, 1.0},
				{4, 1.0},
			},
		},
		{
			name: "Window size 5, beta=1",
			N:    5,
			beta: 1.0,
			checks: []struct {
				index int
				want  float64
			}{
				{0, 1.0 / besselI0(1.0)}, // I0(0)/I0(1) = 1/I0(1) ≈ 0.789848
				{2, 1.0},                 // Центр для нечетного N = 1.0
			},
		},
		{
			name: "Window size 8, beta=2.5",
			N:    8,
			beta: 2.5,
			checks: []struct {
				index int
				want  float64
			}{
				{0, 1.0 / besselI0(2.5)}, // I0(0)/I0(2.5) = 1/I0(2.5)
				// Для четного N нет центрального элемента со значением 1.0
				{7, 1.0 / besselI0(2.5)}, // Симметрично первому
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			window := kaiserWindow(tt.N, tt.beta)

			// Проверяем длину
			if len(window) != tt.N {
				t.Errorf("kaiserWindow(%d, %f) length = %d, want %d",
					tt.N, tt.beta, len(window), tt.N)
			}

			// Проверяем симметричность (для N > 1)
			if tt.N > 1 {
				for i := 0; i < tt.N/2; i++ {
					diff := math.Abs(window[i] - window[tt.N-1-i])
					if diff > 1e-10 {
						t.Errorf("kaiserWindow(%d, %f) not symmetric at indices %d and %d: %f != %f",
							tt.N, tt.beta, i, tt.N-1-i, window[i], window[tt.N-1-i])
					}
				}
			}

			// Проверяем конкретные значения
			for _, check := range tt.checks {
				if check.index >= tt.N {
					continue
				}
				if math.Abs(window[check.index]-check.want) > 1e-6 {
					t.Errorf("kaiserWindow(%d, %f)[%d] = %e, want %e",
						tt.N, tt.beta, check.index, window[check.index], check.want)
				}
			}

			// Проверяем что все значения в диапазоне (0, 1]
			for i, val := range window {
				if val <= 0 || val > 1.0 {
					t.Errorf("kaiserWindow(%d, %f)[%d] = %f, should be in (0, 1]",
						tt.N, tt.beta, i, val)
				}
			}

			// Проверяем что центр = 1.0 для любого beta (для нечетных N)
			if tt.N%2 == 1 && tt.N > 1 {
				center := tt.N / 2
				if math.Abs(window[center]-1.0) > 1e-12 {
					t.Errorf("Center value should be 1.0, got %f", window[center])
				}
			}
		})
	}
}

func TestKaiserWindowApply(t *testing.T) {
	tests := []struct {
		name      string
		coeffs    []float64
		beta      float64
		wantCoeff []float64
		tolerance float64
	}{
		{
			name:      "Empty coefficients",
			coeffs:    []float64{},
			beta:      2.5,
			wantCoeff: []float64{},
			tolerance: 1e-12,
		},
		{
			name:      "Single coefficient, beta=0",
			coeffs:    []float64{2.5},
			beta:      0.0,
			wantCoeff: []float64{2.5 * 1.0}, // Окно из одного элемента = 1.0
			tolerance: 1e-12,
		},
		{
			name:      "All ones, size 3, beta=0",
			coeffs:    []float64{1.0, 1.0, 1.0},
			beta:      0.0,
			wantCoeff: []float64{1.0, 1.0, 1.0}, // Прямоугольное окно
			tolerance: 1e-12,
		},
		{
			name:   "Linear coefficients, size 5, beta=1",
			coeffs: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			beta:   1.0,
			wantCoeff: []float64{
				1.0 * (1.0 / besselI0(1.0)), // 1.0 * (I0(0)/I0(1))
				2.0 * besselI0(1.0*math.Sqrt(1-0.25)) / besselI0(1.0),
				3.0 * 1.0, // Центр для нечетного N = 1.0
				4.0 * besselI0(1.0*math.Sqrt(1-0.25)) / besselI0(1.0),
				5.0 * (1.0 / besselI0(1.0)), // 5.0 * (I0(0)/I0(1))
			},
			tolerance: 1e-12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyKaiserWindow(tt.coeffs, tt.beta)

			if len(got) != len(tt.wantCoeff) {
				t.Errorf("ApplyKaiserWindow() length = %d, want %d", len(got), len(tt.wantCoeff))
				return
			}

			for i := range got {
				if math.Abs(got[i]-tt.wantCoeff[i]) > tt.tolerance {
					t.Errorf("ApplyKaiserWindow()[%d] = %f, want %f (tolerance %e)",
						i, got[i], tt.wantCoeff[i], tt.tolerance)
				}
			}
		})
	}

	// Тест на неизменность исходного массива
	t.Run("Original array not modified", func(t *testing.T) {
		original := []float64{1.0, 2.0, 3.0}
		coeffsCopy := make([]float64, len(original))
		copy(coeffsCopy, original)

		_ = ApplyKaiserWindow(original, 2.5)

		for i := range original {
			if original[i] != coeffsCopy[i] {
				t.Errorf("ApplyKaiserWindow modified original array at index %d: %f != %f",
					i, original[i], coeffsCopy[i])
			}
		}
	})
}
func TestKaiserWindowProperties(t *testing.T) {
	betaValues := []float64{0.0, 1.0, 2.5, 5.0, 10.0}
	NValues := []int{1, 2, 3, 4, 5, 10, 20, 50, 100}

	for _, beta := range betaValues {
		for _, N := range NValues {
			t.Run("N="+string(rune(N))+"_beta="+string(rune(beta)), func(t *testing.T) {
				window := kaiserWindow(N, beta)

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

				// Проверка диапазона значений
				for i, val := range window {
					if val <= 0 || val > 1.0 {
						t.Errorf("Value at %d out of range (0,1]: %f", i, val)
					}
				}

				// Проверка крайних значений
				if N > 1 {
					// Крайние значения должны быть равны I0(0)/I0(beta) = 1/I0(beta)
					expectedEdge := 1.0 / besselI0(beta)
					if math.Abs(window[0]-expectedEdge) > 1e-12 {
						t.Errorf("First value should be %f, got %f", expectedEdge, window[0])
					}
					if math.Abs(window[N-1]-expectedEdge) > 1e-12 {
						t.Errorf("Last value should be %f, got %f", expectedEdge, window[N-1])
					}
				}

				// Проверка центрального значения (для нечетных N)
				if N%2 == 1 && N > 1 {
					center := N / 2
					if math.Abs(window[center]-1.0) > 1e-12 {
						t.Errorf("Center value should be 1.0, got %f", window[center])
					}
				}

				// Проверка что при beta=0 получаем прямоугольное окно
				if beta == 0.0 {
					for i, val := range window {
						if math.Abs(val-1.0) > 1e-12 {
							t.Errorf("For beta=0, all values should be 1.0, got %f at index %d", val, i)
						}
					}
				}

				// Проверка что при увеличении beta крайние значения уменьшаются
				if N > 1 && beta > 0 {
					if window[0] >= 1.0 {
						t.Errorf("For beta>0, edge value should be <1.0, got %f", window[0])
					}
				}
			})
		}
	}
}

func TestKaiserWindowSpecialCases(t *testing.T) {
	// Тест на отрицательное N
	t.Run("Negative N", func(t *testing.T) {
		window := kaiserWindow(-5, 2.5)
		if len(window) != 0 {
			t.Errorf("For negative N, window should be empty, got length %d", len(window))
		}
	})

	// Тест на N=0
	t.Run("Zero N", func(t *testing.T) {
		window := kaiserWindow(0, 2.5)
		if len(window) != 0 {
			t.Errorf("For N=0, window should be empty, got length %d", len(window))
		}
	})

	// Тест на большие beta
	t.Run("Large beta", func(t *testing.T) {
		window := kaiserWindow(10, 100.0)
		// Проверяем что крайние значения очень малы
		if window[0] > 1e-10 {
			t.Errorf("For large beta=100, edge value should be very small, got %e", window[0])
		}
	})
}
