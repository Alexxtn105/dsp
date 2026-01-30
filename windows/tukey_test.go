package windows

import (
	"math"
	"testing"
)

func TestTukeyWindow(t *testing.T) {
	tests := []struct {
		name   string
		N      int
		alpha  float64
		checks []struct {
			index     int
			expected  float64
			tolerance float64
		}
	}{
		{
			name:  "Alpha = 0 (rectangular), N=5",
			N:     5,
			alpha: 0.0,
			checks: []struct {
				index     int
				expected  float64
				tolerance float64
			}{
				{0, 1.0, 1e-12},
				{2, 1.0, 1e-12},
				{4, 1.0, 1e-12},
			},
		},
		{
			name:  "Alpha = 1 (Hann), N=5",
			N:     5,
			alpha: 1.0,
			checks: []struct {
				index     int
				expected  float64
				tolerance float64
			}{
				{0, 0.0, 1e-12},
				{1, 0.5, 1e-12}, // cos^2(pi/4) = 0.5
				{2, 1.0, 1e-12},
				{3, 0.5, 1e-12},
				{4, 0.0, 1e-12},
			},
		},
		{
			name:  "Alpha = 0.5, N=8",
			N:     8,
			alpha: 0.5,
			checks: []struct {
				index     int
				expected  float64
				tolerance float64
			}{
				{0, 0.0, 1e-12},
				{1, 0.611260, 1e-5}, // Было: 0.5
				{2, 1.0, 1e-12},
				{3, 1.0, 1e-12},
				{4, 1.0, 1e-12},
				{5, 1.0, 1e-12},
				{6, 0.611260, 1e-5}, // Было: 0.5
				{7, 0.0, 1e-12},
			},
		},
		{
			name:  "Alpha = 0.3, N=10",
			N:     10,
			alpha: 0.3,
			checks: []struct {
				index     int
				expected  float64
				tolerance float64
			}{
				{0, 0.0, 1e-12},
				{1, 0.843122, 1e-5}, // Было: 0.25, допуск 0.1
				{4, 1.0, 1e-12},
				{8, 0.843122, 1e-5}, // Было: 0.25, допуск 0.1
				{9, 0.0, 1e-12},
			},
		},
		{
			name:  "Single element, any alpha",
			N:     1,
			alpha: 0.5,
			checks: []struct {
				index     int
				expected  float64
				tolerance float64
			}{
				{0, 1.0, 1e-12},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			window := tukeyWindow(tt.N, tt.alpha)

			// Проверяем длину
			if len(window) != tt.N {
				t.Errorf("tukeyWindow(%d, %f) length = %d, want %d", tt.N, tt.alpha, len(window), tt.N)
			}

			// Проверяем симметричность (для N > 1)
			if tt.N > 1 {
				for i := 0; i < tt.N/2; i++ {
					diff := math.Abs(window[i] - window[tt.N-1-i])
					if diff > 1e-10 {
						t.Errorf("tukeyWindow(%d, %f) not symmetric at indices %d and %d: %f != %f",
							tt.N, tt.alpha, i, tt.N-1-i, window[i], window[tt.N-1-i])
					}
				}
			}

			// Проверяем конкретные значения
			for _, check := range tt.checks {
				if check.index >= tt.N {
					continue
				}
				if math.Abs(window[check.index]-check.expected) > check.tolerance {
					t.Errorf("tukeyWindow(%d, %f)[%d] = %f, want %f ± %e",
						tt.N, tt.alpha, check.index, window[check.index], check.expected, check.tolerance)
				}
			}

			// Проверяем что все значения в диапазоне [0, 1]
			for i, val := range window {
				if val < 0 || val > 1 {
					t.Errorf("tukeyWindow(%d, %f)[%d] = %f, should be in [0, 1]", tt.N, tt.alpha, i, val)
				}
			}
		})
	}
}

func TestApplyTukeyWindow(t *testing.T) {
	tests := []struct {
		name      string
		coeffs    []float64
		alpha     float64
		wantCoeff []float64
		tolerance float64
	}{
		{
			name:      "Empty coefficients",
			coeffs:    []float64{},
			alpha:     0.5,
			wantCoeff: []float64{},
			tolerance: 1e-12,
		},
		{
			name:      "Single coefficient, alpha 0.5",
			coeffs:    []float64{2.5},
			alpha:     0.5,
			wantCoeff: []float64{2.5}, // Окно из одного элемента = 1.0
			tolerance: 1e-12,
		},
		{
			name:      "All ones, size 5, alpha 0",
			coeffs:    []float64{1.0, 1.0, 1.0, 1.0, 1.0},
			alpha:     0.0,
			wantCoeff: []float64{1.0, 1.0, 1.0, 1.0, 1.0}, // Прямоугольное окно
			tolerance: 1e-12,
		},
		{
			name:      "All ones, size 5, alpha 1",
			coeffs:    []float64{1.0, 1.0, 1.0, 1.0, 1.0},
			alpha:     1.0,
			wantCoeff: []float64{0.0, 0.5, 1.0, 0.5, 0.0}, // Окно Хэннинга
			tolerance: 1e-12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyTukeyWindow(tt.coeffs, tt.alpha)

			if len(got) != len(tt.wantCoeff) {
				t.Errorf("ApplyTukeyWindow() length = %d, want %d", len(got), len(tt.wantCoeff))
				return
			}

			for i := range got {
				if math.Abs(got[i]-tt.wantCoeff[i]) > tt.tolerance {
					t.Errorf("ApplyTukeyWindow()[%d] = %f, want %f (tolerance %e)",
						i, got[i], tt.wantCoeff[i], tt.tolerance)
				}
			}
		})
	}
}

func TestApplyTukeyWindowDefault(t *testing.T) {
	tests := []struct {
		name      string
		coeffs    []float64
		wantCoeff []float64
		tolerance float64
	}{
		{
			name:      "Default alpha 0.5, size 5",
			coeffs:    []float64{1.0, 1.0, 1.0, 1.0, 1.0},
			wantCoeff: ApplyTukeyWindow([]float64{1.0, 1.0, 1.0, 1.0, 1.0}, 0.5),
			tolerance: 1e-12,
		},
		{
			name:      "Empty coefficients",
			coeffs:    []float64{},
			wantCoeff: []float64{},
			tolerance: 1e-12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyTukeyWindowDefault(tt.coeffs)

			if len(got) != len(tt.wantCoeff) {
				t.Errorf("ApplyTukeyWindowDefault() length = %d, want %d", len(got), len(tt.wantCoeff))
				return
			}

			for i := range got {
				if math.Abs(got[i]-tt.wantCoeff[i]) > tt.tolerance {
					t.Errorf("ApplyTukeyWindowDefault()[%d] = %f, want %f",
						i, got[i], tt.wantCoeff[i])
				}
			}
		})
	}

	// Тест на неизменность исходного массива
	t.Run("Original array not modified", func(t *testing.T) {
		original := []float64{1.0, 2.0, 3.0}
		coeffsCopy := make([]float64, len(original))
		copy(coeffsCopy, original)

		_ = ApplyTukeyWindowDefault(original)

		for i := range original {
			if original[i] != coeffsCopy[i] {
				t.Errorf("ApplyTukeyWindowDefault modified original array at index %d: %f != %f",
					i, original[i], coeffsCopy[i])
			}
		}
	})
}

func TestTukeyWindowEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		N     int
		alpha float64
	}{
		{"Alpha negative", 10, -0.1},
		{"Alpha > 1", 10, 1.1},
		{"N=0", 0, 0.5},
		{"N=2, alpha=0", 2, 0.0},
		{"N=2, alpha=1", 2, 1.0},
		{"N=2, alpha=0.5", 2, 0.5},
		{"N=3, alpha=0.5", 3, 0.5},
		{"N=100, alpha=0.2", 100, 0.2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			window := tukeyWindow(tt.N, tt.alpha)

			// Проверка длины
			if tt.N >= 0 && len(window) != tt.N {
				t.Errorf("Length mismatch: got %d, want %d", len(window), tt.N)
			}

			// Проверка диапазона значений
			for i, val := range window {
				if val < 0 || val > 1 {
					t.Errorf("Value at %d out of range [0,1]: %f", i, val)
				}
			}

			// Проверка симметричности для N > 1
			if tt.N > 1 {
				for i := 0; i < tt.N/2; i++ {
					if math.Abs(window[i]-window[tt.N-1-i]) > 1e-12 {
						t.Errorf("Not symmetric at %d and %d: %f != %f",
							i, tt.N-1-i, window[i], window[tt.N-1-i])
					}
				}
			}
		})
	}
}
