package detectors

import (
	"math"
	"math/cmplx"
)

// FrequencyDetector реализует частотный детектор на основе комплексного умножения
type FrequencyDetector struct {
	sampleRate        float64
	prevSignal        complex128
	prevPhase         float64
	unwrapOffset      float64
	alpha             float64 // Коэффициент сглаживания (для усреднения)
	smoothedFreq      float64 // Текущее сглаженное значение частоты
	smoothInitialized bool    // Флаг инициализации сглаживания
}

// NewFrequencyDetector создает новый частотный детектор
func NewFrequencyDetector(sampleRate float64) *FrequencyDetector {
	if sampleRate <= 0 {
		panic("sampleRate must be positive")
	}

	return &FrequencyDetector{
		sampleRate:        sampleRate,
		prevSignal:        complex(0, 0),
		prevPhase:         math.NaN(),
		unwrapOffset:      0,
		alpha:             0.1, // По умолчанию легкое сглаживание
		smoothedFreq:      0,
		smoothInitialized: false,
	}
}

// SetSmoothingFactor устанавливает коэффициент сглаживания (0.0 - 1.0)
func (fd *FrequencyDetector) SetSmoothingFactor(alpha float64) {
	if alpha < 0 {
		alpha = 0
	} else if alpha > 1 {
		alpha = 1
	}
	fd.alpha = alpha
	// При изменении коэффициента сбрасываем состояние сглаживания
	fd.smoothInitialized = false
}

// DetectFrequency вычисляет мгновенную частоту на основе текущего и предыдущего отсчетов
func (fd *FrequencyDetector) DetectFrequency(signal complex128) float64 {
	if math.IsNaN(fd.prevPhase) {
		// Первый вызов - инициализация
		fd.prevSignal = signal
		fd.prevPhase = fd.calculatePhase(signal)
		return 0
	}

	// Комплексное умножение текущего сигнала на комплексно-сопряженный предыдущий
	// Это дает разность фаз между отсчетами
	phaseDiff := fd.computePhaseDifference(signal)

	// Вычисление мгновенной частоты
	// f = Δφ / (2π * Δt), где Δt = 1/sampleRate
	instantaneousFreq := (phaseDiff * fd.sampleRate) / (2 * math.Pi)

	// Ограничение частоты по теореме Котельникова
	instantaneousFreq = fd.limitFrequency(instantaneousFreq)

	// Применение сглаживания, если нужно
	if fd.alpha > 0 {
		instantaneousFreq = fd.smoothFrequency(instantaneousFreq)
	}

	// Обновление состояния
	fd.prevSignal = signal
	// Фаза обновляется в computePhaseDifference через unwrapPhaseDiff

	return instantaneousFreq
}

// limitFrequency ограничивает частоту диапазоном Найквиста
func (fd *FrequencyDetector) limitFrequency(freq float64) float64 {
	nyquistLimit := fd.sampleRate / 2
	if freq > nyquistLimit {
		return nyquistLimit
	} else if freq < -nyquistLimit {
		return -nyquistLimit
	}
	return freq
}

// ProcessBlock обрабатывает блок данных и возвращает массив мгновенных частот
func (fd *FrequencyDetector) ProcessBlock(signals []complex128) []float64 {
	frequencies := make([]float64, len(signals))

	for i, signal := range signals {
		frequencies[i] = fd.DetectFrequency(signal)
	}

	return frequencies
}

// Reset сбрасывает состояние детектора
func (fd *FrequencyDetector) Reset() {
	fd.prevSignal = complex(0, 0)
	fd.prevPhase = math.NaN()
	fd.unwrapOffset = 0
	fd.smoothedFreq = 0
	fd.smoothInitialized = false
}

// GetInstantaneousPhase возвращает текущую мгновенную фазу
func (fd *FrequencyDetector) GetInstantaneousPhase() float64 {
	if math.IsNaN(fd.prevPhase) {
		return 0
	}
	// Текущая фаза = последняя фаза + накопленное смещение
	currentPhase := fd.prevPhase + fd.unwrapOffset

	// Нормализуем в диапазон [0, 2π)
	currentPhase = math.Mod(currentPhase, 2*math.Pi)
	if currentPhase < 0 {
		currentPhase += 2 * math.Pi
	}

	return currentPhase
}

// GetUnwrapOffset возвращает текущее смещение развертки фазы
func (fd *FrequencyDetector) GetUnwrapOffset() float64 {
	return fd.unwrapOffset
}

// computePhaseDifference вычисляет разность фаз между текущим и предыдущим отсчетом
func (fd *FrequencyDetector) computePhaseDifference(current complex128) float64 {
	// Умножение текущего на комплексно-сопряженный предыдущий
	// conj(prev) * current = |prev||current| * exp(j(φ_current - φ_prev))
	conjugatePrev := complex(real(fd.prevSignal), -imag(fd.prevSignal))
	product := current * conjugatePrev

	// Извлекаем разность фаз из произведения
	phaseDiff := math.Atan2(imag(product), real(product))

	// Развертка фазы для устранения скачков от -π к π
	phaseDiff = fd.unwrapPhaseDiff(phaseDiff)

	return phaseDiff
}

// unwrapPhaseDiff выполняет развертку разности фаз для устранения скачков
func (fd *FrequencyDetector) unwrapPhaseDiff(phaseDiff float64) float64 {
	// Коррекция для скачков через ±π
	if phaseDiff > math.Pi {
		phaseDiff -= 2 * math.Pi
	} else if phaseDiff < -math.Pi {
		phaseDiff += 2 * math.Pi
	}

	// Обновляем смещение для развертки
	fd.unwrapOffset += phaseDiff

	// Обновляем предыдущую фазу для следующего вычисления
	fd.prevPhase += phaseDiff

	// Предотвращаем переполнение (для безопасности)
	fd.preventUnwrapOverflow()

	return phaseDiff
}

// preventUnwrapOverflow предотвращает переполнение unwrapOffset
func (fd *FrequencyDetector) preventUnwrapOverflow() {
	const maxOffset = 1e6 * 2 * math.Pi // Очень большое значение
	if math.Abs(fd.unwrapOffset) > maxOffset {
		// Сбрасываем смещение, сохраняя дробную часть
		fd.unwrapOffset = math.Mod(fd.unwrapOffset, 2*math.Pi)
		// Также корректируем предыдущую фазу
		fd.prevPhase = math.Mod(fd.prevPhase, 2*math.Pi)
	}
}

// calculatePhase вычисляет фазу комплексного числа
func (fd *FrequencyDetector) calculatePhase(signal complex128) float64 {
	return math.Atan2(imag(signal), real(signal))
}

// smoothFrequency применяет экспоненциальное сглаживание к частоте
func (fd *FrequencyDetector) smoothFrequency(currentFreq float64) float64 {
	if !fd.smoothInitialized {
		fd.smoothedFreq = currentFreq
		fd.smoothInitialized = true
		return currentFreq
	}

	fd.smoothedFreq = fd.alpha*currentFreq + (1-fd.alpha)*fd.smoothedFreq
	return fd.smoothedFreq
}

// GetSmoothedFrequency возвращает текущее сглаженное значение частоты
func (fd *FrequencyDetector) GetSmoothedFrequency() float64 {
	if !fd.smoothInitialized {
		return 0
	}
	return fd.smoothedFreq
}

// PLLFrequencyDetector - альтернативная реализация на основе фазовой автоподстройки частоты (PLL)
type PLLFrequencyDetector struct {
	sampleRate float64
	phase      float64
	frequency  float64 // Нормализованная частота (радиан/сэмпл)
	alpha      float64 // Коэффициент петли для фазы
	beta       float64 // Коэффициент петли для частоты
	bandwidth  float64 // Полоса пропускания
}

// NewPLLFrequencyDetector создает частотный детектор на основе PLL
func NewPLLFrequencyDetector(sampleRate, bandwidth float64) *PLLFrequencyDetector {
	if sampleRate <= 0 {
		panic("sampleRate must be positive")
	}
	if bandwidth <= 0 {
		panic("bandwidth must be positive")
	}

	// Расчет коэффициентов для критического затухания
	damping := 0.707                                    // Коэффициент затухания
	naturalFreq := 2 * math.Pi * bandwidth / sampleRate // Нормализованная собственная частота

	// Дискретные коэффициенты для петли 2-го порядка
	alpha := (4 * damping * naturalFreq) /
		(4 + 4*damping*naturalFreq + math.Pow(naturalFreq, 2))
	beta := (4 * math.Pow(naturalFreq, 2)) /
		(4 + 4*damping*naturalFreq + math.Pow(naturalFreq, 2))

	return &PLLFrequencyDetector{
		sampleRate: sampleRate,
		phase:      0,
		frequency:  0,
		alpha:      alpha,
		beta:       beta,
		bandwidth:  bandwidth,
	}
}

// SetBandwidth устанавливает новую полосу пропускания PLL
func (pll *PLLFrequencyDetector) SetBandwidth(bandwidth float64) {
	if bandwidth <= 0 {
		return
	}

	pll.bandwidth = bandwidth
	naturalFreq := 2 * math.Pi * bandwidth / pll.sampleRate
	damping := 0.707

	pll.alpha = (4 * damping * naturalFreq) /
		(4 + 4*damping*naturalFreq + math.Pow(naturalFreq, 2))
	pll.beta = (4 * math.Pow(naturalFreq, 2)) /
		(4 + 4*damping*naturalFreq + math.Pow(naturalFreq, 2))
}

// DetectFrequencyPLL использует PLL для оценки частоты
func (pll *PLLFrequencyDetector) DetectFrequencyPLL(signal complex128) float64 {
	// Нормализация входного сигнала
	magnitude := cmplx.Abs(signal)
	if magnitude > 1e-10 { // Маленький порог для устойчивости
		signal = signal / complex(magnitude, 0)
	} else {
		// Для нулевого или очень малого сигнала используем единичный вектор
		signal = complex(1, 0)
	}

	// Генерация опорного сигнала с текущей фазой
	refSignal := complex(math.Cos(pll.phase), -math.Sin(pll.phase))

	// Фазовый детектор (комплексное умножение)
	phaseDetectorOutput := signal * refSignal

	// Извлечение ошибки фазы
	phaseError := math.Atan2(imag(phaseDetectorOutput), real(phaseDetectorOutput))

	// Обновление фазы и частоты через петлю фильтра
	pll.phase += pll.frequency + pll.alpha*phaseError
	pll.frequency += pll.beta * phaseError

	// Ограничение частоты для устойчивости
	pll.limitNormalizedFrequency()

	// Нормализация фазы
	pll.normalizePhase()

	// Мгновенная частота в Гц
	instantaneousFreq := pll.frequency * pll.sampleRate / (2 * math.Pi)

	// Ограничение по Найквисту
	return pll.limitFrequency(instantaneousFreq)
}

// limitNormalizedFrequency ограничивает нормализованную частоту
func (pll *PLLFrequencyDetector) limitNormalizedFrequency() {
	maxNormalizedFreq := 0.5 // половина частоты дискретизации в нормализованном виде

	if pll.frequency > maxNormalizedFreq {
		pll.frequency = maxNormalizedFreq
	} else if pll.frequency < -maxNormalizedFreq {
		pll.frequency = -maxNormalizedFreq
	}
}

// normalizePhase нормализует фазу в диапазон [0, 2π)
func (pll *PLLFrequencyDetector) normalizePhase() {
	pll.phase = math.Mod(pll.phase, 2*math.Pi)
	if pll.phase < 0 {
		pll.phase += 2 * math.Pi
	}
}

// limitFrequency ограничивает частоту диапазоном Найквиста
func (pll *PLLFrequencyDetector) limitFrequency(freq float64) float64 {
	nyquistLimit := pll.sampleRate / 2
	if freq > nyquistLimit {
		return nyquistLimit
	} else if freq < -nyquistLimit {
		return -nyquistLimit
	}
	return freq
}

// ProcessBlockPLL обрабатывает блок данных с использованием PLL
func (pll *PLLFrequencyDetector) ProcessBlockPLL(signals []complex128) []float64 {
	frequencies := make([]float64, len(signals))

	for i, signal := range signals {
		frequencies[i] = pll.DetectFrequencyPLL(signal)
	}

	return frequencies
}

// ResetPLL сбрасывает состояние PLL детектора
func (pll *PLLFrequencyDetector) ResetPLL() {
	pll.phase = 0
	pll.frequency = 0
}

// GetCurrentPhase возвращает текущую фазу PLL
func (pll *PLLFrequencyDetector) GetCurrentPhase() float64 {
	return pll.phase
}

// GetCurrentNormalizedFrequency возвращает текущую нормализованную частоту
func (pll *PLLFrequencyDetector) GetCurrentNormalizedFrequency() float64 {
	return pll.frequency
}

// GetCurrentFrequency возвращает текущую частоту в Гц
func (pll *PLLFrequencyDetector) GetCurrentFrequency() float64 {
	return pll.frequency * pll.sampleRate / (2 * math.Pi)
}

// FrequencyDetectorConfig конфигурация для создания детектора
type FrequencyDetectorConfig struct {
	SampleRate      float64
	SmoothingFactor float64
	UsePLL          bool
	PLLBandwidth    float64
}

// NewFrequencyDetectorWithConfig создает детектор с конфигурацией
func NewFrequencyDetectorWithConfig(config FrequencyDetectorConfig) interface{} {
	if config.SampleRate <= 0 {
		panic("SampleRate must be positive")
	}

	if config.UsePLL {
		if config.PLLBandwidth <= 0 {
			config.PLLBandwidth = config.SampleRate / 100 // 1% от частоты дискретизации по умолчанию
		}
		return NewPLLFrequencyDetector(config.SampleRate, config.PLLBandwidth)
	} else {
		detector := NewFrequencyDetector(config.SampleRate)
		if config.SmoothingFactor >= 0 {
			detector.SetSmoothingFactor(config.SmoothingFactor)
		}
		return detector
	}
}
