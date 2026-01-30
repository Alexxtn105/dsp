package detectors

import (
	"math"
	"math/cmplx"
	"testing"
)

func TestNewCoherentPhaseDetector(t *testing.T) {
	tests := []struct {
		name           string
		reference      complex128
		alpha          float64
		wantAlpha      float64
		wantRefNormMag float64
	}{
		{
			name:           "нормальные параметры",
			reference:      complex(1, 1),
			alpha:          0.5,
			wantAlpha:      0.5,
			wantRefNormMag: 1.0,
		},
		{
			name:           "альфа меньше нуля -> устанавливается по умолчанию",
			reference:      complex(1, 0),
			alpha:          -0.1,
			wantAlpha:      0.1,
			wantRefNormMag: 1.0,
		},
		{
			name:           "альфа больше единицы -> устанавливается по умолчанию",
			reference:      complex(0, 1),
			alpha:          1.5,
			wantAlpha:      0.1,
			wantRefNormMag: 1.0,
		},
		{
			name:           "нулевой альфа -> устанавливается по умолчанию",
			reference:      complex(1, 1),
			alpha:          0,
			wantAlpha:      0.1,
			wantRefNormMag: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpd := NewCoherentPhaseDetector(tt.reference, tt.alpha)

			if cpd.alpha != tt.wantAlpha {
				t.Errorf("NewCoherentPhaseDetector() alpha = %v, want %v", cpd.alpha, tt.wantAlpha)
			}

			refMagnitude := cmplx.Abs(cpd.referenceSignal)
			if math.Abs(refMagnitude-tt.wantRefNormMag) > 1e-10 {
				t.Errorf("NewCoherentPhaseDetector() reference magnitude = %v, want %v", refMagnitude, tt.wantRefNormMag)
			}

			if cpd.phaseOffset != 0 {
				t.Errorf("NewCoherentPhaseDetector() phaseOffset = %v, want 0", cpd.phaseOffset)
			}

			if cpd.filteredError != 0 {
				t.Errorf("NewCoherentPhaseDetector() filteredError = %v, want 0", cpd.filteredError)
			}
		})
	}
}

func TestNormalizePhase(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{
			name:     "фаза в пределах [-π, π]",
			input:    0.0,
			expected: 0.0,
		},
		{
			name:     "фаза π",
			input:    math.Pi,
			expected: math.Pi, // должно оставаться π
		},
		{
			name:     "фаза -π",
			input:    -math.Pi,
			expected: math.Pi, // -π нормализуется до π
		},
		{
			name:     "фаза больше 2π",
			input:    3 * math.Pi,
			expected: math.Pi,
		},
		{
			name:     "фаза меньше -2π",
			input:    -3 * math.Pi,
			expected: math.Pi, // -3π нормализуется до π
		},
		{
			name:     "фаза 2π",
			input:    2 * math.Pi,
			expected: 0.0,
		},
		{
			name:     "фаза 1.5π",
			input:    1.5 * math.Pi,
			expected: -0.5 * math.Pi,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePhase(tt.input)
			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("normalizePhase(%v) = %v, want %v", tt.input, result, tt.expected)
			}
			// Проверяем, что результат действительно в пределах [-π, π]
			if result < -math.Pi || result > math.Pi {
				t.Errorf("normalizePhase(%v) = %v, not in range [-π, π]", tt.input, result)
			}
		})
	}
}

func TestCoherentPhaseDetector_Detect(t *testing.T) {
	tests := []struct {
		name          string
		reference     complex128
		input         complex128
		alpha         float64
		expectedPhase float64
		tolerance     float64
	}{
		{
			name:          "нулевая разность фаз",
			reference:     cmplx.Exp(complex(0, 0)), // фаза 0
			input:         cmplx.Exp(complex(0, 0)), // фаза 0
			alpha:         1.0,                      // без фильтрации
			expectedPhase: 0.0,
			tolerance:     1e-10,
		},
		{
			name:          "разность фаз π/2",
			reference:     cmplx.Exp(complex(0, 0)),         // фаза 0
			input:         cmplx.Exp(complex(0, math.Pi/2)), // фаза π/2
			alpha:         1.0,
			expectedPhase: math.Pi / 2,
			tolerance:     1e-10,
		},
		{
			name:          "разность фаз -π/2",
			reference:     cmplx.Exp(complex(0, math.Pi/4)),  // фаза π/4
			input:         cmplx.Exp(complex(0, -math.Pi/4)), // фаза -π/4
			alpha:         1.0,
			expectedPhase: -math.Pi / 2,
			tolerance:     1e-10,
		},
		{
			name:          "с фильтрацией (alpha=0.5)",
			reference:     complex(1, 0),
			input:         complex(0, 1), // фаза π/2
			alpha:         0.5,
			expectedPhase: math.Pi / 4, // после первого вызова с alpha=0.5: 0.5 * π/2 + 0.5 * 0 = π/4
			tolerance:     1e-10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpd := NewCoherentPhaseDetector(tt.reference, tt.alpha)
			result := cpd.Detect(tt.input)

			if math.Abs(result-tt.expectedPhase) > tt.tolerance {
				t.Errorf("Detect() = %v, want %v (difference: %v)", result, tt.expectedPhase, math.Abs(result-tt.expectedPhase))
			}
		})
	}
}

func TestCoherentPhaseDetector_UpdateOffset(t *testing.T) {
	cpd := NewCoherentPhaseDetector(complex(1, 0), 1.0)

	// Проверяем начальное состояние
	if cpd.phaseOffset != 0 {
		t.Errorf("initial phaseOffset = %v, want 0", cpd.phaseOffset)
	}

	// Детектируем фазу
	input := complex(0, 1) // фаза π/2
	detected := cpd.Detect(input)

	// Проверяем, что обнаружили правильную фазу
	if math.Abs(detected-math.Pi/2) > 1e-10 {
		t.Errorf("Detect() = %v, want π/2", detected)
	}

	// Обновляем смещение
	cpd.UpdateOffset()

	// Проверяем, что смещение обновилось
	if math.Abs(cpd.phaseOffset-math.Pi/2) > 1e-10 {
		t.Errorf("after UpdateOffset, phaseOffset = %v, want π/2", cpd.phaseOffset)
	}

	// Проверяем, что отфильтрованная ошибка сбросилась
	if cpd.filteredError != 0 {
		t.Errorf("after UpdateOffset, filteredError = %v, want 0", cpd.filteredError)
	}
}

func TestCoherentPhaseDetector_SetPhaseOffset(t *testing.T) {
	cpd := NewCoherentPhaseDetector(complex(1, 0), 0.1)

	tests := []struct {
		name     string
		offset   float64
		expected float64
	}{
		{
			name:     "нулевое смещение",
			offset:   0.0,
			expected: 0.0,
		},
		{
			name:     "смещение π/2",
			offset:   math.Pi / 2,
			expected: math.Pi / 2,
		},
		{
			name:     "смещение больше 2π",
			offset:   3 * math.Pi,
			expected: math.Pi, // должно нормализоваться до π
		},
		{
			name:     "отрицательное смещение",
			offset:   -math.Pi / 2,
			expected: -math.Pi / 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpd.SetPhaseOffset(tt.offset)
			result := cpd.GetPhaseOffset()
			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("SetPhaseOffset(%v) -> GetPhaseOffset() = %v, want %v", tt.offset, result, tt.expected)
			}
		})
	}
}

func TestCoherentPhaseDetector_GetFilteredError(t *testing.T) {
	cpd := NewCoherentPhaseDetector(complex(1, 0), 0.5)

	// Изначально ошибка должна быть 0
	if cpd.GetFilteredError() != 0 {
		t.Errorf("initial GetFilteredError() = %v, want 0", cpd.GetFilteredError())
	}

	// После детектирования ошибка должна измениться
	input := complex(0, 1) // фаза π/2
	cpd.Detect(input)

	// При alpha=0.5 и начальной ошибке 0, после первого измерения:
	// filteredError = 0.5 * (π/2) + 0.5 * 0 = π/4
	expectedError := math.Pi / 4
	result := cpd.GetFilteredError()

	if math.Abs(result-expectedError) > 1e-10 {
		t.Errorf("after Detect(), GetFilteredError() = %v, want %v", result, expectedError)
	}
}

func TestCoherentPhaseDetector_UpdateReferenceSignal(t *testing.T) {
	cpd := NewCoherentPhaseDetector(complex(1, 0), 0.1)

	// Проверяем начальный опорный сигнал
	initialPhase := cmplx.Phase(cpd.referenceSignal)
	if math.Abs(initialPhase) > 1e-10 {
		t.Errorf("initial reference signal phase = %v, want 0", initialPhase)
	}

	// Обновляем опорный сигнал
	newRef := complex(0, 1) // фаза π/2
	cpd.UpdateReferenceSignal(newRef)

	// Проверяем новый опорный сигнал
	newPhase := cmplx.Phase(cpd.referenceSignal)
	expectedPhase := math.Pi / 2
	if math.Abs(newPhase-expectedPhase) > 1e-10 {
		t.Errorf("after UpdateReferenceSignal(), reference phase = %v, want %v", newPhase, expectedPhase)
	}

	// Проверяем, что магнитуда нормирована
	magnitude := cmplx.Abs(cpd.referenceSignal)
	if math.Abs(magnitude-1.0) > 1e-10 {
		t.Errorf("after UpdateReferenceSignal(), reference magnitude = %v, want 1", magnitude)
	}
}

func TestCoherentPhaseDetector_DetectWithOffset(t *testing.T) {
	cpd := NewCoherentPhaseDetector(complex(1, 0), 1.0)

	// Устанавливаем смещение
	cpd.SetPhaseOffset(math.Pi / 4)

	// Детектируем сигнал с фазой π/2
	input := complex(0, 1) // фаза π/2
	result := cpd.Detect(input)

	// Ожидаем: (π/2 - π/4) = π/4
	expected := math.Pi / 4
	if math.Abs(result-expected) > 1e-10 {
		t.Errorf("Detect with offset = %v, want %v", result, expected)
	}
}

func TestCoherentPhaseDetector_MultipleDetections(t *testing.T) {
	cpd := NewCoherentPhaseDetector(complex(1, 0), 0.5)

	// Серия детектирований
	inputs := []complex128{
		complex(0, 1),                    // π/2
		complex(-1, 0),                   // π
		complex(0, -1),                   // -π/2
		cmplx.Exp(complex(0, math.Pi/4)), // π/4
	}

	for i, input := range inputs {
		result := cpd.Detect(input)
		t.Logf("Detection %d: input phase = %v, detected = %v, filteredError = %v",
			i, cmplx.Phase(input), result, cpd.GetFilteredError())

		// Проверяем, что результат в пределах [-π, π]
		if result < -math.Pi || result > math.Pi {
			t.Errorf("Detection %d result = %v, not in range [-π, π]", i, result)
		}
	}

	// Проверяем итоговое значение отфильтрованной ошибки
	finalError := cpd.GetFilteredError()
	if finalError < -math.Pi || finalError > math.Pi {
		t.Errorf("Final filteredError = %v, not in range [-π, π]", finalError)
	}
}

func TestCoherentPhaseDetector_DetectAfterUpdateOffset(t *testing.T) {
	cpd := NewCoherentPhaseDetector(complex(1, 0), 0.8)

	// Первое детектирование
	input1 := complex(0, 1) // фаза π/2
	result1 := cpd.Detect(input1)

	// При alpha=0.8: filteredError = 0.8 * (π/2) + 0.2 * 0 = 0.8 * π/2 = 0.4π
	expected1 := 0.4 * math.Pi
	if math.Abs(result1-expected1) > 1e-10 {
		t.Errorf("First Detect() = %v, want %v", result1, expected1)
	}

	// Обновляем смещение - phaseOffset становится равным текущему filteredError
	cpd.UpdateOffset()

	// Проверяем, что смещение обновилось
	if math.Abs(cpd.GetPhaseOffset()-expected1) > 1e-10 {
		t.Errorf("After UpdateOffset, phaseOffset = %v, want %v", cpd.GetPhaseOffset(), expected1)
	}

	// Второе детектирование - должно быть с учетом обновленного смещения
	input2 := complex(1, 1) // фаза π/4
	result2 := cpd.Detect(input2)

	// После UpdateOffset, filteredError сбрасывается в 0
	// Вычисляем: phaseDiff = π/4 - 0 = π/4
	// filteredError = 0.8 * (π/4) + 0.2 * 0 = 0.2π
	// correctedPhase = filteredError - phaseOffset = 0.2π - 0.4π = -0.2π
	expected2 := -0.2 * math.Pi

	if math.Abs(result2-expected2) > 1e-10 {
		t.Errorf("Detect after UpdateOffset = %v, want %v", result2, expected2)
	}
}

func TestCoherentPhaseDetector_FilteringEffect(t *testing.T) {
	// Тест для проверки эффекта фильтрации с разными alpha
	tests := []struct {
		name   string
		alpha  float64
		inputs []complex128
	}{
		{
			name:  "сильная фильтрация (alpha=0.1)",
			alpha: 0.1,
			inputs: []complex128{
				complex(0, 1),  // π/2
				complex(-1, 0), // π
				complex(0, -1), // -π/2
			},
		},
		{
			name:  "слабая фильтрация (alpha=0.9)",
			alpha: 0.9,
			inputs: []complex128{
				complex(0, 1),  // π/2
				complex(-1, 0), // π
				complex(0, -1), // -π/2
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cpd := NewCoherentPhaseDetector(complex(1, 0), tt.alpha)

			var previousError float64
			for i, input := range tt.inputs {
				_ = cpd.Detect(input)
				currentError := cpd.GetFilteredError()

				if i > 0 {
					// При сильной фильтрации изменения должны быть медленнее
					errorChange := math.Abs(currentError - previousError)
					t.Logf("Step %d: error = %v, change = %v", i, currentError, errorChange)
				}

				previousError = currentError
			}
		})
	}
}
