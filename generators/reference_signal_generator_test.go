package generators

import (
	"math"
	"strings"
	"testing"
)

func TestSignalTypeString(t *testing.T) {
	tests := []struct {
		signalType SignalType
		expected   string
	}{
		{Sine, "Синусоида"},
		{Cosine, "Косинусоида"},
		{Square, "Прямоугольный"},
		{Sawtooth, "Пилообразный"},
		{Triangle, "Треугольный"},
		{SignalType(10), "Неизвестный"},
	}

	for _, tt := range tests {
		result := tt.signalType.String()
		if result != tt.expected {
			t.Errorf("SignalType.String() для %v = %v, ожидается %v", tt.signalType, result, tt.expected)
		}
	}
}

func TestNewReferenceSignalGenerator(t *testing.T) {
	gen := NewReferenceSignalGenerator()

	if gen.Frequency != 1000.0 {
		t.Errorf("Default Frequency = %v, ожидается 1000.0", gen.Frequency)
	}
	if gen.SampleRate != 8000.0 {
		t.Errorf("Default SampleRate = %v, ожидается 8000.0", gen.SampleRate)
	}
	if gen.TotalTime != 1.0 {
		t.Errorf("Default TotalTime = %v, ожидается 1.0", gen.TotalTime)
	}
	if gen.Amplitude != 1.0 {
		t.Errorf("Default Amplitude = %v, ожидается 1.0", gen.Amplitude)
	}
	if gen.Phase != 0.0 {
		t.Errorf("Default Phase = %v, ожидается 0.0", gen.Phase)
	}
	if gen.SignalType != Sine {
		t.Errorf("Default SignalType = %v, ожидается Sine", gen.SignalType)
	}
	if gen.DutyCycle != 0.5 {
		t.Errorf("Default DutyCycle = %v, ожидается 0.5", gen.DutyCycle)
	}
}

func TestGenerateSine(t *testing.T) {
	gen := NewReferenceSignalGenerator()
	gen.Frequency = 1.0
	gen.SampleRate = 10.0
	gen.TotalTime = 1.0
	gen.Amplitude = 2.0
	gen.SignalType = Sine

	signal, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() вернула ошибку: %v", err)
	}

	if len(signal) != 10 {
		t.Errorf("Длина сигнала = %v, ожидается 10", len(signal))
	}

	// Проверим несколько значений
	expectedValues := []float64{
		0.0,                // t=0
		1.1755705045849463, // t=0.1
		1.902113032590307,  // t=0.2
		1.9021130325903073, // t=0.3
		1.1755705045849467, // t=0.4
	}

	for i, expected := range expectedValues {
		if math.Abs(signal[i]-expected) > 1e-10 {
			t.Errorf("signal[%d] = %v, ожидается %v", i, signal[i], expected)
		}
	}
}

func TestGenerateCosine(t *testing.T) {
	gen := NewReferenceSignalGenerator()
	gen.Frequency = 1.0
	gen.SampleRate = 10.0
	gen.TotalTime = 1.0
	gen.Amplitude = 1.0
	gen.SignalType = Cosine
	gen.Phase = math.Pi / 2 // Косинус со сдвигом π/2 должен дать -синус

	signal, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() вернула ошибку: %v", err)
	}

	// cos(ωt + π/2) = -sin(ωt)
	expectedSineGen := NewReferenceSignalGenerator()
	expectedSineGen.Frequency = 1.0
	expectedSineGen.SampleRate = 10.0
	expectedSineGen.TotalTime = 1.0
	expectedSineGen.Amplitude = 1.0
	expectedSineGen.SignalType = Sine
	expectedSignal, _ := expectedSineGen.Generate()

	for i := range signal {
		if math.Abs(signal[i]+expectedSignal[i]) > 1e-10 {
			t.Errorf("signal[%d] = %v, ожидается -%v = %v", i, signal[i], expectedSignal[i], -expectedSignal[i])
		}
	}
}

func TestGenerateSquare(t *testing.T) {
	gen := NewReferenceSignalGenerator()
	gen.Frequency = 2.0
	gen.SampleRate = 40.0
	gen.TotalTime = 0.5
	gen.Amplitude = 5.0
	gen.SignalType = Square
	gen.DutyCycle = 0.25 // 25% заполнения

	signal, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() вернула ошибку: %v", err)
	}

	// Проверим количество положительных и отрицательных значений
	positiveCount := 0
	negativeCount := 0

	for _, value := range signal {
		if value > 0 {
			positiveCount++
		} else if value < 0 {
			negativeCount++
		}
	}

	// За один период: 25% времени +Amplitude, 75% времени -Amplitude
	// Всего периодов: 2 Гц * 0.5 с = 1 период
	// Всего отсчетов: 40 Гц * 0.5 с = 20 отсчетов
	// Положительных: 25% от 20 = 5
	if positiveCount != 5 {
		t.Errorf("Количество положительных значений = %v, ожидается 5", positiveCount)
	}
	if negativeCount != 15 {
		t.Errorf("Количество отрицательных значений = %v, ожидается 15", negativeCount)
	}
}

func TestGenerateSawtooth(t *testing.T) {
	gen := NewReferenceSignalGenerator()
	gen.Frequency = 1.0
	gen.SampleRate = 4.0
	gen.TotalTime = 1.0
	gen.Amplitude = 1.0
	gen.SignalType = Sawtooth

	signal, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() вернула ошибку: %v", err)
	}

	// Пилообразный сигнал от -1 до 1
	// За 1 секунду при частоте 1 Гц - один полный цикл
	// 4 отсчета: t=0, 0.25, 0.5, 0.75
	expected := []float64{-1.0, -0.5, 0.0, 0.5}

	for i, exp := range expected {
		if math.Abs(signal[i]-exp) > 1e-10 {
			t.Errorf("signal[%d] = %v, ожидается %v", i, signal[i], exp)
		}
	}
}

func TestGenerateTriangle(t *testing.T) {
	gen := NewReferenceSignalGenerator()
	gen.Frequency = 1.0
	gen.SampleRate = 8.0
	gen.TotalTime = 1.0
	gen.Amplitude = 1.0
	gen.SignalType = Triangle

	signal, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate() вернула ошибку: %v", err)
	}

	// Треугольный сигнал
	// За 1 секунду при частоте 1 Гц - один полный цикл
	// 8 отсчетов: t=0, 0.125, 0.25, 0.375, 0.5, 0.625, 0.75, 0.875
	expected := []float64{
		0.0,  // 0 * 4 = 0
		0.5,  // 0.125 * 4 = 0.5
		1.0,  // 0.25 * 4 = 1.0
		0.5,  // 2 - 0.375*4 = 0.5
		0.0,  // 2 - 0.5*4 = 0
		-0.5, // 2 - 0.625*4 = -0.5
		-1.0, // 2 - 0.75*4 = -1.0
		-0.5, // 0.875*4 - 4 = -0.5
	}

	for i, exp := range expected {
		if math.Abs(signal[i]-exp) > 1e-10 {
			t.Errorf("signal[%d] = %v, ожидается %v", i, signal[i], exp)
		}
	}
}

func TestInfo(t *testing.T) {
	gen := NewReferenceSignalGenerator()
	gen.Frequency = 100.0
	gen.SampleRate = 1000.0
	gen.TotalTime = 2.0
	gen.Amplitude = 2.5
	gen.Phase = math.Pi / 4
	gen.SignalType = Square
	gen.DutyCycle = 0.3

	info := gen.Info()
	if info == "" {
		t.Error("Info() вернула пустую строку")
	}

	// Проверим наличие ключевых слов в информации
	keywords := []string{
		"Тип сигнала",
		"Частота",
		"Частота дискретизации",
		"Длительность",
		"Амплитуда",
		"Начальная фаза",
		"Коэффициент заполнения",
		"Количество отсчётов",
		"Период сигнала",
	}

	for _, keyword := range keywords {
		if !strings.Contains(info, keyword) {
			t.Errorf("Info() не содержит ключевое слово: %s\nПолученная строка:\n%s", keyword, info)
		}
	}

	// Также проверим конкретные значения
	substrings := []string{
		"Прямоугольный", // Тип сигнала
		"100.0 Гц",      // Частота
		"1000.0 Гц",     // Частота дискретизации
		"2.0 с",         // Длительность
		"2.5",           // Амплитуда
		"0.79 рад",      // Начальная фаза (π/4 ≈ 0.785)
		"30.0%",         // Коэффициент заполнения (0.3 * 100)
		"2000",          // Количество отсчётов (2.0 * 1000.0 = 2000)
	}

	for _, substr := range substrings {
		if !strings.Contains(info, substr) {
			t.Errorf("Info() не содержит подстроку: %s\nПолученная строка:\n%s", substr, info)
		}
	}
}

//// Вспомогательная функция для проверки наличия подстроки
//func contains(s, substr string) bool {
//	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && contains(s[1:], substr))
//}
