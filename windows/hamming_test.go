package windows

import (
	"math"
	"testing"
)

func TestHammingWindow(t *testing.T) {
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
				{0, 0.08}, // alpha - beta*cos(0) = 0.54 - 0.46*1 = 0.08
			},
		},
		{
			name: "Window size 5",
			N:    5,
			checks: []struct {
				index int
				want  float64
			}{
				{0, 0.08}, // alpha - beta*cos(0) = 0.54 - 0.46*1 = 0.08
				{2, 1.0},  // Середина: alpha - beta*cos(pi) = 0.54 - 0.46*(-1) = 1.0
				{4, 0.08}, // Симметрично первому
			},
		},
		{
			name: "Window size 8",
			N:    8,
			checks: []struct {
				index int
				want  float64
			}{
				{0, 0.08},     // alpha - beta*cos(0) = 0.08
				{4, 0.954445}, // alpha - beta*cos(2π*4/7) = 0.54 - 0.46*cos(3.5864) ≈ 0.954445
				{7, 0.08},     // Симметрично первому
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			window := hammingWindow(tt.N)

			// Проверяем длину
			if len(window) != tt.N {
				t.Errorf("hammingWindow(%d) length = %d, want %d", tt.N, len(window), tt.N)
			}

			// Проверяем симметричность (для N > 1)
			if tt.N > 1 {
				for i := 0; i < tt.N/2; i++ {
					diff := math.Abs(window[i] - window[tt.N-1-i])
					if diff > 1e-10 {
						t.Errorf("hammingWindow(%d) not symmetric at indices %d and %d: %f != %f",
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
					t.Errorf("hammingWindow(%d)[%d] = %e, want %e",
						tt.N, check.index, window[check.index], check.want)
				}
			}

			// Проверяем что все значения в диапазоне [0, 1]
			for i, val := range window {
				if val < 0 || val > 1 {
					t.Errorf("hammingWindow(%d)[%d] = %f, should be in [0, 1]", tt.N, i, val)
				}
			}
		})
	}
}

func TestHammingWindowApply(t *testing.T) {
	tests := []struct {
		name      string
		coeffs    []float64
		wantCoeff []float64
		tolerance float64
	}{
		{
			name:      "Empty coefficients",
			coeffs:    []float64{},
			wantCoeff: []float64{},
			tolerance: 1e-12,
		},
		{
			name:      "Single coefficient",
			coeffs:    []float64{2.5},
			wantCoeff: []float64{2.5 * 0.08}, // Окно из одного элемента = 0.08
			tolerance: 1e-12,
		},
		{
			name:      "All ones, size 3",
			coeffs:    []float64{1.0, 1.0, 1.0},
			wantCoeff: []float64{0.08, 1.0, 0.08},
			tolerance: 1e-12,
		},
		{
			name:   "Linear coefficients, size 5",
			coeffs: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			wantCoeff: []float64{
				1.0 * 0.08,
				2.0 * (0.54 - 0.46*math.Cos(math.Pi/2)), // alpha - beta*cos(pi/2) = 0.54 - 0.46*0 = 0.54
				3.0 * 1.0,
				4.0 * (0.54 - 0.46*math.Cos(3*math.Pi/2)), // = 0.54
				5.0 * 0.08,
			},
			tolerance: 1e-12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyHammingWindow(tt.coeffs)

			if len(got) != len(tt.wantCoeff) {
				t.Errorf("ApplyHammingWindow() length = %d, want %d", len(got), len(tt.wantCoeff))
				return
			}

			for i := range got {
				if math.Abs(got[i]-tt.wantCoeff[i]) > tt.tolerance {
					t.Errorf("ApplyHammingWindow()[%d] = %f, want %f (tolerance %e)",
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

		_ = ApplyHammingWindow(original)

		for i := range original {
			if original[i] != coeffsCopy[i] {
				t.Errorf("ApplyHammingWindow modified original array at index %d: %f != %f",
					i, original[i], coeffsCopy[i])
			}
		}
	})
}

func TestHammingWindowProperties(t *testing.T) {
	testCases := []int{1, 2, 3, 4, 5, 10, 20, 50, 100}

	for _, N := range testCases {
		t.Run("N="+string(rune(N)), func(t *testing.T) {
			window := hammingWindow(N)

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
				if val < 0.08 || val > 1.0 { // Хэмминг имеет минимальное значение 0.08
					t.Errorf("Value at %d out of range [0.08,1]: %f", i, val)
				}
			}

			// Проверка крайних значений (для N > 1)
			if N > 1 {
				if math.Abs(window[0]-0.08) > 1e-12 {
					t.Errorf("First value should be 0.08, got %f", window[0])
				}
				if math.Abs(window[N-1]-0.08) > 1e-12 {
					t.Errorf("Last value should be 0.08, got %f", window[N-1])
				}
			}

			// Для N=1 крайнее значение будет 0.08
			if N == 1 && math.Abs(window[0]-0.08) > 1e-12 {
				t.Errorf("For N=1, value should be 0.08, got %f", window[0])
			}
		})
	}
}
