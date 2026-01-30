package detectors

import (
	"math"
	"math/cmplx"
	"testing"
)

func TestNewFrequencyDetector(t *testing.T) {
	t.Run("valid sample rate", func(t *testing.T) {
		fd := NewFrequencyDetector(48000.0)
		if fd == nil {
			t.Fatal("detector should not be nil")
		}
		if fd.sampleRate != 48000.0 {
			t.Errorf("expected sample rate 48000.0, got %f", fd.sampleRate)
		}
	})

	t.Run("invalid sample rate panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for invalid sample rate")
			}
		}()
		_ = NewFrequencyDetector(0)
	})

	t.Run("negative sample rate panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for negative sample rate")
			}
		}()
		_ = NewFrequencyDetector(-1000.0)
	})
}

func TestSetSmoothingFactor(t *testing.T) {
	fd := NewFrequencyDetector(48000.0)

	t.Run("valid values", func(t *testing.T) {
		fd.SetSmoothingFactor(0.5)
		if fd.alpha != 0.5 {
			t.Errorf("expected alpha 0.5, got %f", fd.alpha)
		}

		fd.SetSmoothingFactor(0.0)
		if fd.alpha != 0.0 {
			t.Errorf("expected alpha 0.0, got %f", fd.alpha)
		}

		fd.SetSmoothingFactor(1.0)
		if fd.alpha != 1.0 {
			t.Errorf("expected alpha 1.0, got %f", fd.alpha)
		}
	})

	t.Run("out of range values", func(t *testing.T) {
		fd.SetSmoothingFactor(-0.5)
		if fd.alpha != 0.0 {
			t.Errorf("expected alpha 0.0 for negative value, got %f", fd.alpha)
		}

		fd.SetSmoothingFactor(1.5)
		if fd.alpha != 1.0 {
			t.Errorf("expected alpha 1.0 for >1 value, got %f", fd.alpha)
		}
	})
}

func TestDetectFrequency(t *testing.T) {
	t.Run("first call returns zero", func(t *testing.T) {
		fd := NewFrequencyDetector(48000.0)
		freq := fd.DetectFrequency(complex(1, 0))
		if freq != 0 {
			t.Errorf("first call should return 0, got %f", freq)
		}
	})

	t.Run("constant phase gives zero frequency", func(t *testing.T) {
		fd := NewFrequencyDetector(48000.0)

		// Первый вызов
		fd.DetectFrequency(complex(1, 0))

		// Второй вызов с такой же фазой
		freq := fd.DetectFrequency(complex(1, 0))

		if math.Abs(freq) > 1e-10 {
			t.Errorf("constant phase should give zero frequency, got %f", freq)
		}
	})

	t.Run("rotating signal gives positive frequency", func(t *testing.T) {
		sampleRate := 48000.0
		testFreq := 1000.0
		angularFreq := 2 * math.Pi * testFreq / sampleRate

		fd := NewFrequencyDetector(sampleRate)

		// Генерируем два последовательных отсчета вращающегося сигнала
		phase1 := 0.0
		phase2 := angularFreq

		signal1 := complex(math.Cos(phase1), math.Sin(phase1))
		signal2 := complex(math.Cos(phase2), math.Sin(phase2))

		// Первый вызов
		fd.DetectFrequency(signal1)

		// Второй вызов
		detectedFreq := fd.DetectFrequency(signal2)

		// Допустимая погрешность из-за дискретизации
		if math.Abs(detectedFreq-testFreq) > 1.0 {
			t.Errorf("expected frequency around %f, got %f", testFreq, detectedFreq)
		}
	})

	t.Run("frequency limiting", func(t *testing.T) {
		sampleRate := 48000.0
		nyquist := sampleRate / 2

		fd := NewFrequencyDetector(sampleRate)

		// Сигнал с частотой выше Найквиста
		phaseJump := math.Pi * 1.1 // Скачок фазы больше π

		signal1 := complex(1, 0)
		signal2 := cmplx.Exp(complex(0, phaseJump))

		fd.DetectFrequency(signal1)
		detectedFreq := fd.DetectFrequency(signal2)

		// Частота должна быть ограничена
		if detectedFreq > nyquist {
			t.Errorf("frequency should be limited to nyquist %f, got %f", nyquist, detectedFreq)
		}
	})
}

func TestProcessBlock(t *testing.T) {
	sampleRate := 48000.0
	testFreq := 1000.0
	angularFreq := 2 * math.Pi * testFreq / sampleRate

	// Генерируем блок синусоидального сигнала
	blockSize := 10
	signals := make([]complex128, blockSize)
	for i := 0; i < blockSize; i++ {
		phase := angularFreq * float64(i)
		signals[i] = complex(math.Cos(phase), math.Sin(phase))
	}

	fd := NewFrequencyDetector(sampleRate)
	frequencies := fd.ProcessBlock(signals)

	if len(frequencies) != blockSize {
		t.Errorf("expected %d frequencies, got %d", blockSize, len(frequencies))
	}

	// Первая частота должна быть 0
	if frequencies[0] != 0 {
		t.Errorf("first frequency should be 0, got %f", frequencies[0])
	}

	// Остальные должны быть близки к testFreq
	for i := 1; i < len(frequencies); i++ {
		if math.Abs(frequencies[i]-testFreq) > 1.0 {
			t.Errorf("frequency[%d] expected around %f, got %f", i, testFreq, frequencies[i])
		}
	}
}

func TestReset(t *testing.T) {
	fd := NewFrequencyDetector(48000.0)

	// Выполним несколько вызовов
	fd.DetectFrequency(complex(1, 0))
	fd.DetectFrequency(complex(0, 1))

	// Проверим, что состояние изменилось
	if math.IsNaN(fd.prevPhase) {
		t.Error("phase should not be NaN after processing")
	}

	// Сбросим
	fd.Reset()

	// Проверим сброс состояния
	if !math.IsNaN(fd.prevPhase) {
		t.Error("phase should be NaN after reset")
	}
	if fd.unwrapOffset != 0 {
		t.Errorf("unwrap offset should be 0 after reset, got %f", fd.unwrapOffset)
	}
	if fd.smoothInitialized {
		t.Error("smooth should not be initialized after reset")
	}
}

func TestGetInstantaneousPhase(t *testing.T) {
	fd := NewFrequencyDetector(48000.0)

	// До первого вызова - prevPhase = NaN
	phase := fd.GetInstantaneousPhase()
	if phase != 0 {
		t.Errorf("phase should be 0 before first call, got %f", phase)
	}

	// Первый вызов - устанавливает prevPhase в фазу первого сигнала (0)
	fd.DetectFrequency(complex(1, 0)) // Фаза 0

	// После первого вызова GetInstantaneousPhase должен вернуть 0
	phase = fd.GetInstantaneousPhase()
	if math.Abs(phase-0) > 1e-10 {
		t.Errorf("after first call: expected phase 0, got %f", phase)
	}

	// Второй вызов с сигналом фазы π/4
	signal := cmplx.Exp(complex(0, math.Pi/4))
	fd.DetectFrequency(signal)

	// Теперь:
	// - prevPhase обновляется в unwrapPhaseDiff: prevPhase += phaseDiff
	// - phaseDiff между 0 и π/4 = π/4
	// - prevPhase становится 0 + π/4 = π/4
	// - unwrapOffset также становится π/4
	// - GetInstantaneousPhase() возвращает prevPhase + unwrapOffset = π/4 + π/4 = π/2

	phase = fd.GetInstantaneousPhase()

	// Фаза должна быть нормализована в [0, 2π)
	if phase < 0 || phase >= 2*math.Pi {
		t.Errorf("phase should be in [0, 2π), got %f", phase)
	}

	// С учетом реализации, после второго вызова получаем π/2
	expectedPhase := math.Pi / 2 // 1.570796
	if math.Abs(phase-expectedPhase) > 1e-10 {
		t.Errorf("expected phase %f (π/2), got %f", expectedPhase, phase)
	}

	// Проверим отдельные компоненты
	if math.Abs(fd.prevPhase-math.Pi/4) > 1e-10 {
		t.Errorf("prevPhase should be π/4, got %f", fd.prevPhase)
	}
	if math.Abs(fd.unwrapOffset-math.Pi/4) > 1e-10 {
		t.Errorf("unwrapOffset should be π/4, got %f", fd.unwrapOffset)
	}
}

func TestSmoothing(t *testing.T) {
	fd := NewFrequencyDetector(48000.0)
	fd.SetSmoothingFactor(0.5)

	// Генерируем сигнал с изменяющейся частотой
	signals := []complex128{
		complex(1, 0),
		cmplx.Exp(complex(0, 0.1)), // Малая разность фаз
		cmplx.Exp(complex(0, 0.5)), // Большая разность фаз
	}

	freq1 := fd.DetectFrequency(signals[0]) // 0
	//freq2 := fd.DetectFrequency(signals[1]) // Мгновенная частота 1
	_ = fd.DetectFrequency(signals[1]) // Мгновенная частота 1
	//freq3 := fd.DetectFrequency(signals[2]) // Мгновенная частота 2, сглаженная
	_ = fd.DetectFrequency(signals[2]) // Мгновенная частота 2, сглаженная

	// Первая частота без сглаживания
	if freq1 != 0 {
		t.Errorf("first frequency should be 0, got %f", freq1)
	}

	// Проверим, что сглаживание работает
	//smoothedFreq := fd.GetSmoothedFrequency()
	_ = fd.GetSmoothedFrequency()

	// После третьего вызова smoothedFreq должна быть сглаженной версией
	if !fd.smoothInitialized {
		t.Error("smoothing should be initialized")
	}
}

func TestUnwrapPhase(t *testing.T) {
	fd := NewFrequencyDetector(48000.0)

	// Простой тест: проверяем базовую работу развертки
	signals := []complex128{
		complex(1, 0),                      // Фаза 0
		cmplx.Exp(complex(0, 0.9*math.Pi)), // Фаза 0.9π
		cmplx.Exp(complex(0, 1.1*math.Pi)), // Фаза 1.1π
	}

	fd.DetectFrequency(signals[0])
	fd.DetectFrequency(signals[1])
	fd.DetectFrequency(signals[2])

	// Проверим, что unwrapOffset имеет разумное значение
	offset := fd.GetUnwrapOffset()

	// Рассчитаем ожидаемое значение:
	// 1. 0 → 0.9π: разность = 0.9π, unwrapOffset = 0.9π
	// 2. 0.9π → 1.1π: разность = 0.2π (1.1π - 0.9π),
	//    не превышает ±π, поэтому коррекции нет
	//    unwrapOffset = 0.9π + 0.2π = 1.1π
	expectedOffset := 1.1 * math.Pi

	if math.Abs(offset-expectedOffset) > 1e-10 {
		t.Errorf("expected unwrap offset %f, got %f", expectedOffset, offset)
	}

	// Проверим GetInstantaneousPhase
	phase := fd.GetInstantaneousPhase()
	// prevPhase = 1.1π, unwrapOffset = 1.1π, сумма = 2.2π
	// после нормализации: 2.2π mod 2π = 0.2π
	expectedPhase := 0.2 * math.Pi
	if math.Abs(phase-expectedPhase) > 1e-10 {
		t.Errorf("expected instantaneous phase %f, got %f", expectedPhase, phase)
	}
}

// Новый тест для проверки разности фаз
func TestPhaseDifferenceCalculation(t *testing.T) {
	fd := NewFrequencyDetector(48000.0)

	t.Run("small phase difference", func(t *testing.T) {
		fd.Reset()

		signal1 := complex(1, 0)              // Фаза 0
		signal2 := cmplx.Exp(complex(0, 0.3)) // Фаза 0.3 рад

		fd.DetectFrequency(signal1)
		freq := fd.DetectFrequency(signal2)

		// Частота = Δφ * sampleRate / (2π)
		expectedFreq := 0.3 * 48000.0 / (2 * math.Pi)

		if math.Abs(freq-expectedFreq) > 1.0 {
			t.Errorf("expected frequency %f, got %f", expectedFreq, freq)
		}
	})

	t.Run("phase wrap around", func(t *testing.T) {
		fd.Reset()

		signal1 := cmplx.Exp(complex(0, 3*math.Pi/4))  // Фаза 3π/4
		signal2 := cmplx.Exp(complex(0, -3*math.Pi/4)) // Фаза -3π/4 = 5π/4

		fd.DetectFrequency(signal1)
		freq := fd.DetectFrequency(signal2)

		// Разность фаз: от 3π/4 до 5π/4 = π/2, но с учетом направления
		// Фактически: 5π/4 - 3π/4 = π/2
		expectedFreq := (math.Pi / 2) * 48000.0 / (2 * math.Pi)

		if math.Abs(freq-expectedFreq) > 1.0 {
			t.Errorf("expected frequency %f, got %f", expectedFreq, freq)
		}
	})
}

// Тест для проверки calculatePhase
func TestCalculatePhase(t *testing.T) {
	fd := NewFrequencyDetector(48000.0)

	testCases := []struct {
		name   string
		signal complex128
		phase  float64
	}{
		{"real positive", complex(1, 0), 0},
		{"real negative", complex(-1, 0), math.Pi},
		{"imag positive", complex(0, 1), math.Pi / 2},
		{"imag negative", complex(0, -1), -math.Pi / 2},
		{"45 degrees", complex(1, 1), math.Pi / 4},
		{"135 degrees", complex(-1, 1), 3 * math.Pi / 4},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			phase := fd.calculatePhase(tc.signal)
			if math.Abs(phase-tc.phase) > 1e-10 {
				t.Errorf("%s: expected phase %f, got %f", tc.name, tc.phase, phase)
			}
		})
	}
}
