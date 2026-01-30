package detectors

import (
	"math"
	"testing"
)

func TestNewPLLFrequencyDetector(t *testing.T) {
	t.Run("valid parameters", func(t *testing.T) {
		pll := NewPLLFrequencyDetector(48000.0, 1000.0)
		if pll == nil {
			t.Fatal("detector should not be nil")
		}
		if pll.sampleRate != 48000.0 {
			t.Errorf("expected sample rate 48000.0, got %f", pll.sampleRate)
		}
		if pll.bandwidth != 1000.0 {
			t.Errorf("expected bandwidth 1000.0, got %f", pll.bandwidth)
		}
	})

	t.Run("invalid sample rate panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for invalid sample rate")
			}
		}()
		_ = NewPLLFrequencyDetector(0, 1000.0)
	})

	t.Run("invalid bandwidth panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for invalid bandwidth")
			}
		}()
		_ = NewPLLFrequencyDetector(48000.0, 0)
	})
}

func TestSetBandwidth(t *testing.T) {
	pll := NewPLLFrequencyDetector(48000.0, 1000.0)

	originalAlpha := pll.alpha
	originalBeta := pll.beta

	// Устанавливаем новую полосу пропускания
	newBandwidth := 2000.0
	pll.SetBandwidth(newBandwidth)

	if pll.bandwidth != newBandwidth {
		t.Errorf("expected bandwidth %f, got %f", newBandwidth, pll.bandwidth)
	}

	// Коэффициенты должны измениться
	if pll.alpha == originalAlpha || pll.beta == originalBeta {
		t.Error("loop coefficients should change with bandwidth")
	}

	// Проверим, что коэффициент damping = 0.707 сохраняется
	// (встроено в реализацию)

	t.Run("invalid bandwidth ignored", func(t *testing.T) {
		pll := NewPLLFrequencyDetector(48000.0, 1000.0)
		originalBandwidth := pll.bandwidth

		pll.SetBandwidth(0)
		if pll.bandwidth != originalBandwidth {
			t.Errorf("bandwidth should not change for invalid value, got %f", pll.bandwidth)
		}

		pll.SetBandwidth(-100)
		if pll.bandwidth != originalBandwidth {
			t.Errorf("bandwidth should not change for negative value, got %f", pll.bandwidth)
		}
	})
}

func TestDetectFrequencyPLL(t *testing.T) {
	sampleRate := 48000.0
	bandwidth := 1000.0
	testFreq := 2000.0

	pll := NewPLLFrequencyDetector(sampleRate, bandwidth)

	// Генерируем синусоидальный сигнал
	angularFreq := 2 * math.Pi * testFreq / sampleRate
	blockSize := 150

	var lastFreq float64
	for i := 0; i < blockSize; i++ {
		phase := angularFreq * float64(i)
		signal := complex(math.Cos(phase), math.Sin(phase))

		detectedFreq := pll.DetectFrequencyPLL(signal)
		lastFreq = detectedFreq

		// На начальных итерациях частота может не сойтись
		if i > blockSize/2 {
			// Во второй половине PLL должен захватить сигнал
			if math.Abs(detectedFreq-testFreq) > bandwidth {
				t.Errorf("iteration %d: frequency error too large: expected ~%f, got %f",
					i, testFreq, detectedFreq)
			}
		}
	}

	// В конце PLL должен точно отслеживать частоту
	if math.Abs(lastFreq-testFreq) > 10.0 { // Допустимая погрешность 10 Гц
		t.Errorf("final frequency error too large: expected ~%f, got %f",
			testFreq, lastFreq)
	}
}

func TestDetectFrequencyPLLWithSmallSignal(t *testing.T) {
	pll := NewPLLFrequencyDetector(48000.0, 1000.0)

	// Очень маленький сигнал
	tinySignal := complex(1e-20, 1e-20)

	// Не должно быть паники или NaN
	freq := pll.DetectFrequencyPLL(tinySignal)

	if math.IsNaN(freq) || math.IsInf(freq, 0) {
		t.Errorf("frequency should be finite for small signal, got %f", freq)
	}
}

func TestLimitNormalizedFrequency(t *testing.T) {
	pll := NewPLLFrequencyDetector(48000.0, 1000.0)

	// Искусственно установим слишком большую частоту
	pll.frequency = 0.6 // > 0.5 (Найквист в нормализованном виде)

	pll.limitNormalizedFrequency()

	if pll.frequency > 0.5 {
		t.Errorf("normalized frequency should be limited to 0.5, got %f", pll.frequency)
	}

	// Проверим отрицательную частоту
	pll.frequency = -0.6
	pll.limitNormalizedFrequency()

	if pll.frequency < -0.5 {
		t.Errorf("normalized frequency should be limited to -0.5, got %f", pll.frequency)
	}
}

func TestPLLNormalizePhase(t *testing.T) {
	pll := NewPLLFrequencyDetector(48000.0, 1000.0)

	testCases := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"already normalized", math.Pi, math.Pi},
		{"greater than 2π", 3 * math.Pi, math.Pi},
		{"less than 0", -math.Pi / 2, 3 * math.Pi / 2},
		{"much greater", 7 * math.Pi, math.Pi},
		{"much less", -5 * math.Pi / 2, 3 * math.Pi / 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pll.phase = tc.input
			pll.normalizePhase()

			if math.Abs(pll.phase-tc.expected) > 1e-10 {
				t.Errorf("expected %f, got %f", tc.expected, pll.phase)
			}

			// Дополнительно проверяем диапазон
			if pll.phase < 0 || pll.phase >= 2*math.Pi {
				t.Errorf("phase %f not in [0, 2π)", pll.phase)
			}
		})
	}
}

func TestProcessBlockPLL(t *testing.T) {
	sampleRate := 48000.0
	bandwidth := 1000.0
	testFreq := 1500.0

	// Генерируем блок сигнала
	blockSize := 50
	signals := make([]complex128, blockSize)
	angularFreq := 2 * math.Pi * testFreq / sampleRate

	for i := 0; i < blockSize; i++ {
		phase := angularFreq * float64(i)
		signals[i] = complex(math.Cos(phase), math.Sin(phase))
	}

	pll := NewPLLFrequencyDetector(sampleRate, bandwidth)
	frequencies := pll.ProcessBlockPLL(signals)

	if len(frequencies) != blockSize {
		t.Errorf("expected %d frequencies, got %d", blockSize, len(frequencies))
	}

	// Проверим, что PLL схватывает частоту
	// Последние несколько значений должны быть близки к testFreq
	lastCount := 10
	if blockSize > lastCount {
		sum := 0.0
		for i := blockSize - lastCount; i < blockSize; i++ {
			sum += frequencies[i]
		}
		avgFreq := sum / float64(lastCount)

		if math.Abs(avgFreq-testFreq) > bandwidth/2 {
			t.Errorf("PLL did not lock: average frequency %f, expected %f", avgFreq, testFreq)
		}
	}
}

func TestResetPLL(t *testing.T) {
	pll := NewPLLFrequencyDetector(48000.0, 1000.0)

	// Выполним несколько вызовов
	signal := complex(math.Cos(0.1), math.Sin(0.1))
	pll.DetectFrequencyPLL(signal)
	pll.DetectFrequencyPLL(signal)

	// Проверим, что состояние изменилось
	if pll.phase == 0 || pll.frequency == 0 {
		t.Error("PLL state should change after processing")
	}

	// Сбросим
	pll.ResetPLL()

	// Проверим сброс состояния
	if pll.phase != 0 {
		t.Errorf("phase should be 0 after reset, got %f", pll.phase)
	}
	if pll.frequency != 0 {
		t.Errorf("frequency should be 0 after reset, got %f", pll.frequency)
	}
}

func TestGetCurrentMethods(t *testing.T) {
	sampleRate := 48000.0
	pll := NewPLLFrequencyDetector(sampleRate, 1000.0)

	// Инициализируем PLL с известной частотой
	pll.frequency = 0.1 // Нормализованная частота
	pll.phase = math.Pi / 2

	// Проверим GetCurrentPhase
	currentPhase := pll.GetCurrentPhase()
	if currentPhase != math.Pi/2 {
		t.Errorf("expected phase π/2, got %f", currentPhase)
	}

	// Проверим GetCurrentNormalizedFrequency
	normFreq := pll.GetCurrentNormalizedFrequency()
	if normFreq != 0.1 {
		t.Errorf("expected normalized frequency 0.1, got %f", normFreq)
	}

	// Проверим GetCurrentFrequency
	expectedFreq := 0.1 * sampleRate / (2 * math.Pi)
	currentFreq := pll.GetCurrentFrequency()

	if math.Abs(currentFreq-expectedFreq) > 1e-10 {
		t.Errorf("expected frequency %f, got %f", expectedFreq, currentFreq)
	}
}

func TestNewFrequencyDetectorWithConfig(t *testing.T) {
	t.Run("create regular detector", func(t *testing.T) {
		config := FrequencyDetectorConfig{
			SampleRate:      48000.0,
			SmoothingFactor: 0.5,
			UsePLL:          false,
		}

		detector := NewFrequencyDetectorWithConfig(config)

		fd, ok := detector.(*FrequencyDetector)
		if !ok {
			t.Fatal("expected *FrequencyDetector")
		}

		if fd.sampleRate != 48000.0 {
			t.Errorf("expected sample rate 48000.0, got %f", fd.sampleRate)
		}

		// Проверим косвенно, что сглаживание установлено
		if fd.alpha != 0.5 {
			t.Errorf("expected smoothing factor 0.5, got %f", fd.alpha)
		}
	})

	t.Run("create PLL detector", func(t *testing.T) {
		config := FrequencyDetectorConfig{
			SampleRate:   44100.0,
			UsePLL:       true,
			PLLBandwidth: 2000.0,
		}

		detector := NewFrequencyDetectorWithConfig(config)

		pll, ok := detector.(*PLLFrequencyDetector)
		if !ok {
			t.Fatal("expected *PLLFrequencyDetector")
		}

		if pll.sampleRate != 44100.0 {
			t.Errorf("expected sample rate 44100.0, got %f", pll.sampleRate)
		}

		if pll.bandwidth != 2000.0 {
			t.Errorf("expected bandwidth 2000.0, got %f", pll.bandwidth)
		}
	})

	t.Run("PLL detector with default bandwidth", func(t *testing.T) {
		config := FrequencyDetectorConfig{
			SampleRate: 96000.0,
			UsePLL:     true,
			// PLLBandwidth не указан
		}

		detector := NewFrequencyDetectorWithConfig(config)
		pll := detector.(*PLLFrequencyDetector)

		expectedBandwidth := 96000.0 / 100 // 1% по умолчанию
		if math.Abs(pll.bandwidth-expectedBandwidth) > 1e-10 {
			t.Errorf("expected default bandwidth %f, got %f", expectedBandwidth, pll.bandwidth)
		}
	})

	t.Run("invalid sample rate panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for invalid sample rate")
			}
		}()

		config := FrequencyDetectorConfig{
			SampleRate: 0,
			UsePLL:     false,
		}

		_ = NewFrequencyDetectorWithConfig(config)
	})
}
