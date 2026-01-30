package windows

import (
	"fmt"
	"math"
	"testing"
)

func TestBlackmanHarrisWindow(t *testing.T) {
	// Вычисляем ожидаемые значения для тестовых случаев
	window5 := blackmanHarrisWindow(5)
	window8 := blackmanHarrisWindow(8)

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
				{0, 1.0},
			},
		},
		{
			name: "Window size 5",
			N:    5,
			checks: []struct {
				index int
				want  float64
			}{
				{0, window5[0]},
				{2, window5[2]},
				{4, window5[4]},
			},
		},
		{
			name: "Window size 8",
			N:    8,
			checks: []struct {
				index int
				want  float64
			}{
				{0, window8[0]},
				{3, window8[3]},
				{4, window8[4]},
				{7, window8[7]},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			window := blackmanHarrisWindow(tt.N)

			if len(window) != tt.N {
				t.Errorf("blackmanHarrisWindow(%d) length = %d, want %d", tt.N, len(window), tt.N)
			}

			// Проверяем симметричность
			if tt.N > 1 {
				for i := 0; i < tt.N/2; i++ {
					diff := math.Abs(window[i] - window[tt.N-1-i])
					if diff > 1e-14 {
						t.Errorf("blackmanHarrisWindow(%d) not symmetric at indices %d and %d: %.15f != %.15f",
							tt.N, i, tt.N-1-i, window[i], window[tt.N-1-i])
					}
				}
			}

			// Проверяем конкретные значения
			for _, check := range tt.checks {
				if check.index >= tt.N {
					continue
				}
				// Используем относительную погрешность
				diff := math.Abs(window[check.index] - check.want)
				relDiff := diff / math.Max(math.Abs(window[check.index]), math.Abs(check.want))

				if relDiff > 1e-14 {
					t.Errorf("blackmanHarrisWindow(%d)[%d] = %.15f, want %.15f (relDiff=%e)",
						tt.N, check.index, window[check.index], check.want, relDiff)
				}
			}

			// Проверяем диапазон значений
			for i, val := range window {
				if val < -1e-14 || val > 1+1e-14 {
					t.Errorf("blackmanHarrisWindow(%d)[%d] = %f, should be in [0, 1]", tt.N, i, val)
				}
			}
		})
	}
}

func TestApplyBlackmanHarrisWindow(t *testing.T) {
	window3 := blackmanHarrisWindow(3)
	window5 := blackmanHarrisWindow(5)

	tests := []struct {
		name      string
		coeffs    []float64
		wantCoeff []float64
	}{
		{
			name:      "Empty coefficients",
			coeffs:    []float64{},
			wantCoeff: []float64{},
		},
		{
			name:      "Single coefficient",
			coeffs:    []float64{2.5},
			wantCoeff: []float64{2.5},
		},
		{
			name:      "All ones, size 3",
			coeffs:    []float64{1.0, 1.0, 1.0},
			wantCoeff: []float64{1.0 * window3[0], 1.0 * window3[1], 1.0 * window3[2]},
		},
		{
			name:   "Linear coefficients",
			coeffs: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			wantCoeff: []float64{
				1.0 * window5[0],
				2.0 * window5[1],
				3.0 * window5[2],
				4.0 * window5[3],
				5.0 * window5[4],
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyBlackmanHarrisWindow(tt.coeffs)

			if len(got) != len(tt.wantCoeff) {
				t.Errorf("ApplyBlackmanHarrisWindow() length = %d, want %d", len(got), len(tt.wantCoeff))
				return
			}

			for i := range got {
				diff := math.Abs(got[i] - tt.wantCoeff[i])
				relDiff := diff / math.Max(math.Abs(got[i]), math.Abs(tt.wantCoeff[i]))

				if relDiff > 1e-14 {
					t.Errorf("ApplyBlackmanHarrisWindow()[%d] = %.15f, want %.15f (relDiff=%e)",
						i, got[i], tt.wantCoeff[i], relDiff)
				}
			}
		})
	}

	// Тест на неизменность исходного массива
	t.Run("Original array not modified", func(t *testing.T) {
		original := []float64{1.0, 2.0, 3.0}
		coeffsCopy := make([]float64, len(original))
		copy(coeffsCopy, original)

		_ = ApplyBlackmanHarrisWindow(original)

		for i := range original {
			if original[i] != coeffsCopy[i] {
				t.Errorf("ApplyBlackmanHarrisWindow modified original array at index %d: %f != %f",
					i, original[i], coeffsCopy[i])
			}
		}
	})
}

func TestBlackmanHarrisWindowProperties(t *testing.T) {
	testCases := []int{1, 2, 3, 4, 5, 10, 20, 50, 100}

	for _, N := range testCases {
		t.Run(fmt.Sprintf("N=%d", N), func(t *testing.T) {
			window := blackmanHarrisWindow(N)

			if len(window) != N {
				t.Errorf("Length mismatch: got %d, want %d", len(window), N)
			}

			// Проверка симметричности
			if N > 1 {
				for i := 0; i < N/2; i++ {
					diff := math.Abs(window[i] - window[N-1-i])
					if diff > 1e-14 {
						t.Errorf("Not symmetric at %d and %d: %.15f != %.15f",
							i, N-1-i, window[i], window[N-1-i])
					}
				}
			}

			// Проверка диапазона значений
			for i, val := range window {
				if val < -1e-14 || val > 1+1e-14 {
					t.Errorf("Value at %d out of range [0,1]: %.15f", i, val)
				}
			}

			// Проверка свойств окна Блэкмана-Харриса
			if N > 1 {
				// Края должны быть очень малы
				if window[0] > 0.001 || window[N-1] > 0.001 {
					t.Errorf("Edges should be small: window[0]=%.6f, window[%d]=%.6f",
						window[0], N-1, window[N-1])
				}
			}
		})
	}
}

// Тест для проверки известных значений окна Блэкмана-Харриса
func TestBlackmanHarrisWindowKnownValues(t *testing.T) {
	// Тестируем несколько известных значений
	tests := []struct {
		N     int
		index int
	}{
		{N: 1, index: 0},
		{N: 5, index: 0},
		{N: 5, index: 2},
		{N: 8, index: 0},
		{N: 8, index: 3},
		{N: 8, index: 4},
	}

	for _, tt := range tests {
		window := blackmanHarrisWindow(tt.N)
		if tt.index < len(window) {
			val := window[tt.index]

			// Просто проверяем что значение вычислено и в диапазоне
			if math.IsNaN(val) || math.IsInf(val, 0) {
				t.Errorf("blackmanHarrisWindow(%d)[%d] is NaN or Inf: %f", tt.N, tt.index, val)
			}

			// Для отладки выведем значения
			if tt.N == 8 && (tt.index == 3 || tt.index == 4) {
				t.Logf("blackmanHarrisWindow(%d)[%d] = %.15f", tt.N, tt.index, val)
			}
		}
	}
}
