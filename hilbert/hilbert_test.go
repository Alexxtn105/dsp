package hilbert

import (
	"math"
	"math/cmplx"
	"testing"
)

// Тест проверки основных свойств преобразования Гильберта
func TestHilbertTransform(t *testing.T) {
	sampleRate := 48000.0
	order := 127

	ht := NewHilbertTransform(sampleRate, order)

	// Проверяем, что порядок нечетный
	if ht.order%2 == 0 {
		t.Errorf("Order should be odd, got %d", ht.order)
	}

	// Проверяем групповую задержку
	expectedDelay := order / 2
	if ht.GetGroupDelay() != expectedDelay {
		t.Errorf("Expected group delay %d, got %d",
			expectedDelay, ht.GetGroupDelay())
	}
}

// Тест на синусоидальном сигнале
func TestHilbertSineWave(t *testing.T) {
	sampleRate := 48000.0
	order := 127
	frequency := 1000.0

	ht := NewHilbertTransform(sampleRate, order)

	// Пропускаем переходный процесс
	warmupSamples := order * 2

	for i := 0; i < warmupSamples; i++ {
		t := float64(i) / sampleRate
		input := math.Sin(2.0 * math.Pi * frequency * t)
		ht.Tick(input)
	}

	// Проверяем установившийся режим
	var sumMagnitude float64
	testSamples := 1000

	for i := 0; i < testSamples; i++ {
		t := float64(warmupSamples+i) / sampleRate
		input := math.Sin(2.0 * math.Pi * frequency * t)
		output := ht.Tick(input)

		magnitude := cmplx.Abs(output)
		sumMagnitude += magnitude
	}

	avgMagnitude := sumMagnitude / float64(testSamples)

	// Для синусоиды амплитудой 1.0, аналитический сигнал должен иметь
	// амплитуду близкую к 1.0
	expectedMagnitude := 1.0
	tolerance := 0.1

	if math.Abs(avgMagnitude-expectedMagnitude) > tolerance {
		t.Errorf("Expected magnitude ≈ %.2f, got %.4f",
			expectedMagnitude, avgMagnitude)
	}
}

// Тест сброса состояния
func TestHilbertReset(t *testing.T) {
	sampleRate := 48000.0
	order := 63

	ht := NewHilbertTransform(sampleRate, order)

	// Обрабатываем несколько отсчетов
	for i := 0; i < 100; i++ {
		ht.Tick(0.5)
	}

	// Сбрасываем
	ht.Reset()

	// Проверяем, что буфер обнулен
	output := ht.Tick(0.0)
	if real(output) != 0.0 || imag(output) != 0.0 {
		t.Errorf("After reset, output should be 0+0i, got %v", output)
	}
}

// Бенчмарк производительности
func BenchmarkHilbertTransform(b *testing.B) {
	sampleRate := 48000.0
	order := 127

	ht := NewHilbertTransform(sampleRate, order)
	input := 0.5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ht.Tick(input)
	}
}
