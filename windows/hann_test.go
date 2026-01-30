package windows

import (
	"math"
	"testing"
)

func TestHannWindow(t *testing.T) {
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
				{0, 0.0}, // При N=1: alpha - alpha*cos(0) = 0.5 - 0.5*1 = 0.0
			},
		},
		{
			name: "Window size 5",
			N:    5,
			checks: []struct {
				index int
				want  float64
			}{
				{0, 0.0}, // alpha - alpha*cos(0) = 0.5 - 0.5*1 = 0.0
				{2, 1.0}, // Середина: alpha - alpha*cos(pi) = 0.5 - 0.5*(-1) = 1.0
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
				{0, 0.0},
				{4, 0.5 - 0.5*math.Cos(2*math.Pi*4/7)}, // Исправляем ожидаемое значение
				{7, 0.0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			window := hannWindow(tt.N)

			// Проверяем длину
			if len(window) != tt.N {
				t.Errorf("hannWindow(%d) length = %d, want %d", tt.N, len(window), tt.N)
			}

			// Проверяем симметричность (для N > 1)
			if tt.N > 1 {
				for i := 0; i < tt.N/2; i++ {
					diff := math.Abs(window[i] - window[tt.N-1-i])
					if diff > 1e-10 {
						t.Errorf("hannWindow(%d) not symmetric at indices %d and %d: %f != %f",
							tt.N, i, tt.N-1-i, window[i], window[tt.N-1-i])
					}
				}
			}

			// Проверяем конкретные значения
			for _, check := range tt.checks {
				if check.index >= tt.N {
					continue
				}
				if math.Abs(window[check.index]-check.want) > 1e-12 {
					t.Errorf("hannWindow(%d)[%d] = %e, want %e",
						tt.N, check.index, window[check.index], check.want)
				}
			}

			// Проверяем что все значения в диапазоне [0, 1]
			for i, val := range window {
				if val < 0 || val > 1 {
					t.Errorf("hannWindow(%d)[%d] = %f, should be in [0, 1]", tt.N, i, val)
				}
			}
		})
	}
}

func TestApplyHannWindow(t *testing.T) {
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
			wantCoeff: []float64{0.0}, // Окно из одного элемента = 0.0
			tolerance: 1e-12,
		},
		{
			name:      "All ones, size 3",
			coeffs:    []float64{1.0, 1.0, 1.0},
			wantCoeff: []float64{0.0, 1.0, 0.0},
			tolerance: 1e-12,
		},
		{
			name:   "Linear coefficients, size 5",
			coeffs: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			wantCoeff: []float64{
				1.0 * 0.0,
				2.0 * 0.5, // alpha - alpha*cos(pi/2) = 0.5 - 0.5*0 = 0.5
				3.0 * 1.0,
				4.0 * 0.5,
				5.0 * 0.0,
			},
			tolerance: 1e-12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyHannWindow(tt.coeffs)

			if len(got) != len(tt.wantCoeff) {
				t.Errorf("ApplyHannWindow() length = %d, want %d", len(got), len(tt.wantCoeff))
				return
			}

			for i := range got {
				if math.Abs(got[i]-tt.wantCoeff[i]) > tt.tolerance {
					t.Errorf("ApplyHannWindow()[%d] = %f, want %f (tolerance %e)",
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

		_ = ApplyHannWindow(original)

		for i := range original {
			if original[i] != coeffsCopy[i] {
				t.Errorf("ApplyHannWindow modified original array at index %d: %f != %f",
					i, original[i], coeffsCopy[i])
			}
		}
	})
}

func TestHannWindowProperties(t *testing.T) {
	testCases := []int{1, 2, 3, 4, 5, 10, 20, 50, 100}

	for _, N := range testCases {
		t.Run("N="+string(rune(N)), func(t *testing.T) {
			window := hannWindow(N)

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
				if val < 0 || val > 1 {
					t.Errorf("Value at %d out of range [0,1]: %f", i, val)
				}
			}

			// Проверка крайних значений (должны быть 0 для N > 1)
			if N > 1 {
				if math.Abs(window[0]) > 1e-12 {
					t.Errorf("First value should be 0, got %f", window[0])
				}
				if math.Abs(window[N-1]) > 1e-12 {
					t.Errorf("Last value should be 0, got %f", window[N-1])
				}
			}
		})
	}
}
