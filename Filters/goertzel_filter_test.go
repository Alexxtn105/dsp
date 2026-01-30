package filters

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

// Тест обработки сигнала
func TestGoertzelFilter_Process(t *testing.T) {
	// Создаем чистый синусоидальный сигнал
	createSineWave := func(freq, samplingRate float64, n int, amplitude, phase float64) []float64 {
		signal := make([]float64, n)
		for i := 0; i < n; i++ {
			signal[i] = amplitude * math.Sin(2*math.Pi*freq*float64(i)/samplingRate+phase)
		}
		return signal
	}

	tests := []struct {
		name         string
		freq         float64
		samplingRate float64
		totalN       int
		amplitude    float64
		phase        float64
	}{
		{
			name:         "1000 Hz signal at 8000 Hz sampling",
			freq:         1000,
			samplingRate: 8000,
			totalN:       256,
			amplitude:    1.0,
			phase:        0,
		},
		{
			name:         "500 Hz signal at 44100 Hz sampling",
			freq:         516.796875, // 12 * 44100 / 1024
			samplingRate: 44100,
			totalN:       1024,
			amplitude:    1.0,
			phase:        math.Pi / 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем фильтр для целевой частоты
			filter, err := NewGoertzelFilter(tt.freq, tt.samplingRate, tt.totalN)
			if err != nil {
				t.Fatalf("failed to create filter: %v", err)
			}

			// Создаем тестовый сигнал с целевой частотой
			signal := createSineWave(tt.freq, tt.samplingRate, tt.totalN, tt.amplitude, tt.phase)
			deltaF := tt.samplingRate / float64(tt.totalN)
			t.Logf("Parameters: k=%d, w=%v, cosW=%v, sinW=%v", filter.GetCoefficient(), filter.w, filter.cosW, filter.sinW)
			t.Logf("Processed %d samples out of %d", filter.GetProcessedCount(), tt.totalN)
			t.Logf("delta_f=%v", deltaF) //= 44100 / 1024 ≈ 43.07 Гц

			// Обрабатываем все отсчеты
			for _, sample := range signal {
				if err := filter.Process(sample); err != nil {
					t.Fatalf("failed to process sample: %v", err)
				}
			}

			// Проверяем, что все отсчеты обработаны
			if !filter.IsComplete() {
				t.Errorf("filter should be complete after processing all samples")
			}

			// Получаем амплитуду
			magnitude, err := filter.GetMagnitude()
			if err != nil {
				t.Fatalf("failed to get magnitude: %v", err)
			}

			// Проверяем, что амплитуда близка к ожидаемой
			// Для синусоидального сигнала амплитуда должна быть близка к исходной
			expectedMagnitude := tt.amplitude
			tolerance := 0.01 // 1% допуск

			diff := math.Abs(magnitude - expectedMagnitude)
			if diff > tolerance*expectedMagnitude {
				t.Errorf("magnitude = %v, want %v ± %v (diff: %v)",
					magnitude, expectedMagnitude, tolerance*expectedMagnitude, diff)
			}

			// Также проверяем оптимизированную версию
			magnitudeOpt, err := filter.GetMagnitudeOptimized()
			if err != nil {
				t.Fatalf("failed to get optimized magnitude: %v", err)
			}

			diffOpt := math.Abs(magnitudeOpt - expectedMagnitude)
			if diffOpt > tolerance*expectedMagnitude {
				t.Errorf("optimized magnitude = %v, want %v ± %v (diff: %v)",
					magnitudeOpt, expectedMagnitude, tolerance*expectedMagnitude, diffOpt)
			}

			t.Logf("Target frequency: %v Hz", tt.freq)
			t.Logf("Detected magnitude: %v", magnitude)
			t.Logf("Optimized magnitude: %v", magnitudeOpt)
			t.Logf("Expected magnitude: %v", expectedMagnitude)
			t.Logf("Difference: %v", diff)
		})
	}
}

// Тест обработки нескольких сигналов
func TestGoertzelFilter_MultipleSignals(t *testing.T) {
	samplingRate := 8000.0
	totalN := 256

	// Фильтр для 1000 Гц
	filter1000, err := NewGoertzelFilter(1000, samplingRate, totalN)
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	// Фильтр для 2000 Гц
	filter2000, err := NewGoertzelFilter(2000, samplingRate, totalN)
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	// Фильтр для 3000 Гц (сигнала нет)
	filter3000, err := NewGoertzelFilter(3000, samplingRate, totalN)
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	// Создаем сигнал с двумя частотами
	signal := make([]float64, totalN)
	for i := 0; i < totalN; i++ {
		// 1000 Гц с амплитудой 1.0
		signal[i] = math.Sin(2*math.Pi*1000*float64(i)/samplingRate) +
			// 2000 Гц с амплитудой 0.5
			0.5*math.Sin(2*math.Pi*2000*float64(i)/samplingRate)
	}

	// Обрабатываем сигнал всеми фильтрами
	for i := 0; i < totalN; i++ {
		filter1000.Process(signal[i])
		filter2000.Process(signal[i])
		filter3000.Process(signal[i])
	}

	// Проверяем результаты
	mag1000, _ := filter1000.GetMagnitude()
	mag2000, _ := filter2000.GetMagnitude()
	mag3000, _ := filter3000.GetMagnitude()

	// 1000 Гц должна иметь наибольшую амплитуду (~1.0)
	if mag1000 < 0.95 {
		t.Errorf("1000 Hz magnitude (%v) should be close to 1.0", mag1000)
	}

	// 2000 Гц должна иметь амплитуду около 0.5
	if math.Abs(mag2000-0.5) > 0.05 {
		t.Errorf("2000 Hz magnitude (%v) should be close to 0.5", mag2000)
	}

	// 3000 Гц должна иметь очень маленькую амплитуду (шум)
	if mag3000 > 0.05 {
		t.Errorf("3000 Hz magnitude (%v) should be very small (< 0.05) since it's not in the signal", mag3000)
	}

	t.Logf("1000 Hz magnitude: %v (expected: ~1.0)", mag1000)
	t.Logf("2000 Hz magnitude: %v (expected: ~0.5)", mag2000)
	t.Logf("3000 Hz magnitude: %v (expected: < 0.05)", mag3000)
}

// Тест с шумом
func TestGoertzelFilter_WithNoise(t *testing.T) {
	// Используем фиксированный seed для воспроизводимости
	rand.Seed(42)

	filter, err := NewGoertzelFilter(1000, 8000, 512)
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	// Создаем сигнал с шумом
	signal := make([]float64, 512)
	for i := 0; i < 512; i++ {
		// Чистый сигнал 1000 Гц с амплитудой 1.0
		signal[i] = math.Sin(2*math.Pi*1000*float64(i)/8000) +
			// Белый шум с амплитудой 0.2
			0.2*(rand.Float64()-0.5)
	}

	// Обрабатываем сигнал
	for _, sample := range signal {
		filter.Process(sample)
	}

	magnitude, err := filter.GetMagnitude()
	if err != nil {
		t.Fatalf("failed to get magnitude: %v", err)
	}

	// Амплитуда должна быть примерно 1.0 плюс/минус влияние шума
	expected := 1.0
	tolerance := 0.15 // 15% допуск из-за шума

	if math.Abs(magnitude-expected) > tolerance {
		t.Errorf("magnitude with noise = %v, want %v ± %v", magnitude, expected, tolerance)
	}

	t.Logf("Magnitude with noise: %v", magnitude)
	t.Logf("Expected: %v ± %v", expected, tolerance)
}

// Тест с различными амплитудами
func TestGoertzelFilter_VariousAmplitudes(t *testing.T) {
	samplingRate := 8000.0
	totalN := 256
	freq := 1000.0

	amplitudes := []float64{0.1, 0.5, 1.0, 2.0, 5.0}
	tolerance := 0.02 // 2% допуск

	for _, amplitude := range amplitudes {
		t.Run(fmt.Sprintf("amplitude_%.1f", amplitude), func(t *testing.T) {
			filter, err := NewGoertzelFilter(freq, samplingRate, totalN)
			if err != nil {
				t.Fatalf("failed to create filter: %v", err)
			}

			// Создаем сигнал
			signal := make([]float64, totalN)
			for i := 0; i < totalN; i++ {
				signal[i] = amplitude * math.Sin(2*math.Pi*freq*float64(i)/samplingRate)
			}

			// Обрабатываем
			for _, sample := range signal {
				filter.Process(sample)
			}

			magnitude, err := filter.GetMagnitude()
			if err != nil {
				t.Fatalf("failed to get magnitude: %v", err)
			}

			expected := amplitude
			diff := math.Abs(magnitude - expected)

			if diff > tolerance*expected {
				t.Errorf("amplitude %.1f: magnitude = %v, want %v ± %v (diff: %v)",
					amplitude, magnitude, expected, tolerance*expected, diff)
			}

			t.Logf("Input amplitude: %.1f", amplitude)
			t.Logf("Detected magnitude: %v", magnitude)
			t.Logf("Expected: %v", expected)
		})
	}
}

// Тест с нулевой амплитудой (шум)
func TestGoertzelFilter_ZeroAmplitude(t *testing.T) {
	samplingRate := 8000.0
	totalN := 256
	freq := 1000.0

	filter, err := NewGoertzelFilter(freq, samplingRate, totalN)
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	// Создаем сигнал только с шумом (без целевой частоты)
	rand.Seed(42)
	signal := make([]float64, totalN)
	for i := 0; i < totalN; i++ {
		signal[i] = 0.1 * (rand.Float64() - 0.5) // Только шум
	}

	// Обрабатываем
	for _, sample := range signal {
		filter.Process(sample)
	}

	magnitude, err := filter.GetMagnitude()
	if err != nil {
		t.Fatalf("failed to get magnitude: %v", err)
	}

	// Амплитуда должна быть очень маленькой
	if magnitude > 0.05 {
		t.Errorf("magnitude for noise-only signal = %v, should be < 0.05", magnitude)
	}

	t.Logf("Noise-only signal magnitude: %v (should be < 0.05)", magnitude)
}

// Тест проверки согласованности двух методов расчета
func TestGoertzelFilter_MethodsConsistency(t *testing.T) {
	samplingRate := 8000.0
	totalN := 128
	freq := 1500.0

	filter, err := NewGoertzelFilter(freq, samplingRate, totalN)
	if err != nil {
		t.Fatalf("failed to create filter: %v", err)
	}

	// Создаем сложный сигнал
	signal := make([]float64, totalN)
	for i := 0; i < totalN; i++ {
		signal[i] = math.Sin(2*math.Pi*1500*float64(i)/samplingRate) +
			0.3*math.Sin(2*math.Pi*2500*float64(i)/samplingRate) +
			0.1*math.Sin(2*math.Pi*3500*float64(i)/samplingRate)
	}

	// Обрабатываем
	for _, sample := range signal {
		filter.Process(sample)
	}

	magnitude1, _ := filter.GetMagnitude()
	magnitude2, _ := filter.GetMagnitudeOptimized()
	power, _ := filter.GetPower()

	// Проверяем согласованность методов
	diff := math.Abs(magnitude1 - magnitude2)
	if diff > 1e-10 {
		t.Errorf("methods inconsistent: magnitude1 = %v, magnitude2 = %v, diff = %v",
			magnitude1, magnitude2, diff)
	}

	// Проверяем связь между амплитудой и мощностью
	expectedPower := magnitude1 * magnitude1 / 2
	powerDiff := math.Abs(power - expectedPower)
	if powerDiff > 1e-10 {
		t.Errorf("power calculation inconsistent: power = %v, expected = %v, diff = %v",
			power, expectedPower, powerDiff)
	}

	t.Logf("Method 1 magnitude: %v", magnitude1)
	t.Logf("Method 2 magnitude: %v", magnitude2)
	t.Logf("Power: %v", power)
	t.Logf("Methods difference: %v", diff)
}

// Вспомогательная функция
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}

func BenchmarkSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = i * i
	}
}
