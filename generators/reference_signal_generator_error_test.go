package generators

import (
	"math"
	"strings"
	"testing"
)

func TestValidateErrors(t *testing.T) {
	tests := []struct {
		name        string
		modifyGen   func(*ReferenceSignalGenerator)
		expectError bool
		errorSubstr string
	}{
		{
			name: "Valid parameters",
			modifyGen: func(gen *ReferenceSignalGenerator) {
				// Все параметры по умолчанию - валидны
			},
			expectError: false,
		},
		{
			name: "Negative frequency",
			modifyGen: func(gen *ReferenceSignalGenerator) {
				gen.Frequency = -100.0
			},
			expectError: true,
			errorSubstr: "частота должна быть положительной",
		},
		{
			name: "Zero frequency",
			modifyGen: func(gen *ReferenceSignalGenerator) {
				gen.Frequency = 0.0
			},
			expectError: true,
			errorSubstr: "частота должна быть положительной",
		},
		{
			name: "Negative sample rate",
			modifyGen: func(gen *ReferenceSignalGenerator) {
				gen.SampleRate = -8000.0
			},
			expectError: true,
			errorSubstr: "частота дискретизации должна быть положительной",
		},
		{
			name: "Zero sample rate",
			modifyGen: func(gen *ReferenceSignalGenerator) {
				gen.SampleRate = 0.0
			},
			expectError: true,
			errorSubstr: "частота дискретизации должна быть положительной",
		},
		{
			name: "Negative total time",
			modifyGen: func(gen *ReferenceSignalGenerator) {
				gen.TotalTime = -1.0
			},
			expectError: true,
			errorSubstr: "длительность должна быть положительной",
		},
		{
			name: "Zero total time",
			modifyGen: func(gen *ReferenceSignalGenerator) {
				gen.TotalTime = 0.0
			},
			expectError: true,
			errorSubstr: "длительность должна быть положительной",
		},
		{
			name: "Negative amplitude",
			modifyGen: func(gen *ReferenceSignalGenerator) {
				gen.Amplitude = -1.0
			},
			expectError: true,
			errorSubstr: "амплитуда должна быть положительной",
		},
		{
			name: "Zero amplitude",
			modifyGen: func(gen *ReferenceSignalGenerator) {
				gen.Amplitude = 0.0
			},
			expectError: true,
			errorSubstr: "амплитуда должна быть положительной",
		},
		{
			name: "Duty cycle zero",
			modifyGen: func(gen *ReferenceSignalGenerator) {
				gen.DutyCycle = 0.0
			},
			expectError: true,
			errorSubstr: "коэффициент заполнения должен быть в диапазоне (0, 1)",
		},
		{
			name: "Duty cycle one",
			modifyGen: func(gen *ReferenceSignalGenerator) {
				gen.DutyCycle = 1.0
			},
			expectError: true,
			errorSubstr: "коэффициент заполнения должен быть в диапазоне (0, 1)",
		},
		{
			name: "Duty cycle negative",
			modifyGen: func(gen *ReferenceSignalGenerator) {
				gen.DutyCycle = -0.1
			},
			expectError: true,
			errorSubstr: "коэффициент заполнения должен быть в диапазоне (0, 1)",
		},
		{
			name: "Duty cycle greater than one",
			modifyGen: func(gen *ReferenceSignalGenerator) {
				gen.DutyCycle = 1.1
			},
			expectError: true,
			errorSubstr: "коэффициент заполнения должен быть в диапазоне (0, 1)",
		},
		{
			name: "Nyquist violation",
			modifyGen: func(gen *ReferenceSignalGenerator) {
				gen.Frequency = 5000.0
				gen.SampleRate = 8000.0
			},
			expectError: true,
			errorSubstr: "нарушен критерий Найквиста",
		},
		{
			name: "Nyquist boundary case",
			modifyGen: func(gen *ReferenceSignalGenerator) {
				gen.Frequency = 4000.0
				gen.SampleRate = 8000.0
			},
			expectError: true,
			errorSubstr: "нарушен критерий Найквиста",
		},
		{
			name: "Valid Nyquist case",
			modifyGen: func(gen *ReferenceSignalGenerator) {
				gen.Frequency = 3999.0
				gen.SampleRate = 8000.0
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewReferenceSignalGenerator()
			tt.modifyGen(gen)

			_, err := gen.Generate()

			if tt.expectError {
				if err == nil {
					t.Error("Ожидалась ошибка, но её нет")
				} else if tt.errorSubstr != "" && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("Ошибка должна содержать '%s', получено: '%s'", tt.errorSubstr, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Не ожидалась ошибка, но получено: %v", err)
				}
			}
		})
	}
}

func TestGenerateWithDifferentParameters(t *testing.T) {
	tests := []struct {
		name     string
		freq     float64
		sr       float64
		time     float64
		expected int
	}{
		{"1 second, 10 Hz", 1.0, 10.0, 1.0, 10},
		{"0.5 second, 20 Hz", 1.0, 20.0, 0.5, 10},
		{"2 seconds, 5 Hz", 1.0, 5.0, 2.0, 10},
		{"0.1 second, 100 Hz", 1.0, 100.0, 0.1, 10},
		{"Round test", 1.0, 44.1, 0.1, 4}, // 44.1 * 0.1 = 4.41 → округление до 4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewReferenceSignalGenerator()
			gen.Frequency = tt.freq
			gen.SampleRate = tt.sr
			gen.TotalTime = tt.time

			signal, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate() вернула ошибку: %v", err)
			}

			if len(signal) != tt.expected {
				t.Errorf("Длина сигнала = %v, ожидается %v", len(signal), tt.expected)
			}
		})
	}
}

func TestPhaseShift(t *testing.T) {
	gen1 := NewReferenceSignalGenerator()
	gen1.Frequency = 1.0
	gen1.SampleRate = 10.0
	gen1.TotalTime = 1.0
	gen1.Amplitude = 1.0
	gen1.SignalType = Sine
	gen1.Phase = 0.0

	gen2 := NewReferenceSignalGenerator()
	gen2.Frequency = 1.0
	gen2.SampleRate = 10.0
	gen2.TotalTime = 1.0
	gen2.Amplitude = 1.0
	gen2.SignalType = Sine
	gen2.Phase = math.Pi / 2 // Сдвиг на 90 градусов

	_, err1 := gen1.Generate()
	if err1 != nil {
		t.Fatalf("Generate() вернула ошибку: %v", err1)
	}

	signal2, err2 := gen2.Generate()
	if err2 != nil {
		t.Fatalf("Generate() вернула ошибку: %v", err2)
	}

	// sin(ωt + π/2) = cos(ωt)
	// Проверим, что signal2 соответствует косинусу
	genCosine := NewReferenceSignalGenerator()
	genCosine.Frequency = 1.0
	genCosine.SampleRate = 10.0
	genCosine.TotalTime = 1.0
	genCosine.Amplitude = 1.0
	genCosine.SignalType = Cosine
	genCosine.Phase = 0.0
	cosSignal, _ := genCosine.Generate()

	for i := range signal2 {
		if math.Abs(signal2[i]-cosSignal[i]) > 1e-10 {
			t.Errorf("Сигнал со сдвигом фазы π/2 не соответствует косинусу в точке %d: %v != %v", i, signal2[i], cosSignal[i])
		}
	}
}
