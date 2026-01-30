package generators

import (
	"fmt"
	"math"
)

// SignalType определяет тип генерируемого сигнала
type SignalType int

const (
	Sine     SignalType = iota // Синусоидальный сигнал
	Cosine                     // Косинусоидальный сигнал
	Square                     // Прямоугольный сигнал
	Sawtooth                   // Пилообразный сигнал
	Triangle                   // Треугольный сигнал
)

// String возвращает строковое представление типа сигнала
func (st SignalType) String() string {
	switch st {
	case Sine:
		return "Синусоида"
	case Cosine:
		return "Косинусоида"
	case Square:
		return "Прямоугольный"
	case Sawtooth:
		return "Пилообразный"
	case Triangle:
		return "Треугольный"
	default:
		return "Неизвестный"
	}
}

// ReferenceSignalGenerator генерирует эталонный сигнал
type ReferenceSignalGenerator struct {
	Frequency  float64    // Частота сигнала в герцах
	SampleRate float64    // Частота дискретизации в герцах
	TotalTime  float64    // Длительность сигнала в секундах
	Amplitude  float64    // Амплитуда сигнала
	Phase      float64    // Начальная фаза в радианах
	SignalType SignalType // Тип сигнала
	DutyCycle  float64    // Коэффициент заполнения (0.0 - 1.0) для прямоугольного сигнала
}

// NewReferenceSignalGenerator создает новый генератор с настройками по умолчанию
func NewReferenceSignalGenerator() *ReferenceSignalGenerator {
	return &ReferenceSignalGenerator{
		Frequency:  1000.0,
		SampleRate: 8000.0,
		TotalTime:  1.0,
		Amplitude:  1.0,
		Phase:      0.0,
		SignalType: Sine,
		DutyCycle:  0.5, // 50% заполнение по умолчанию
	}
}

// Generate создает массив отсчётов сигнала
func (rsg *ReferenceSignalGenerator) Generate() ([]float64, error) {
	// Проверка входных параметров
	if err := rsg.validate(); err != nil {
		return nil, err
	}

	numSamples := int(math.Round(rsg.TotalTime * rsg.SampleRate))
	signals := make([]float64, numSamples)

	// Предвычисление констант для оптимизации
	timeStep := 1.0 / rsg.SampleRate
	angularFreq := 2 * math.Pi * rsg.Frequency

	// Генерация сигнала в зависимости от типа
	for i := 0; i < numSamples; i++ {
		time := float64(i) * timeStep
		normalizedTime := rsg.Frequency * time // Время, нормированное на период

		switch rsg.SignalType {
		case Sine:
			signals[i] = rsg.generateSine(angularFreq, time)
		case Cosine:
			signals[i] = rsg.generateCosine(angularFreq, time)
		case Square:
			signals[i] = rsg.generateSquare(normalizedTime)
		case Sawtooth:
			signals[i] = rsg.generateSawtooth(normalizedTime)
		case Triangle:
			signals[i] = rsg.generateTriangle(normalizedTime)
		}
	}

	return signals, nil
}

// generateSine генерирует синусоидальный сигнал
func (rsg *ReferenceSignalGenerator) generateSine(angularFreq, time float64) float64 {
	return rsg.Amplitude * math.Sin(angularFreq*time+rsg.Phase)
}

// generateCosine генерирует косинусоидальный сигнал
func (rsg *ReferenceSignalGenerator) generateCosine(angularFreq, time float64) float64 {
	return rsg.Amplitude * math.Cos(angularFreq*time+rsg.Phase)
}

// generateSquare генерирует прямоугольный сигнал
func (rsg *ReferenceSignalGenerator) generateSquare(normalizedTime float64) float64 {
	// Фаза с учетом начальной фазы
	phase := normalizedTime + rsg.Phase/(2*math.Pi)
	fractionalPart := phase - math.Floor(phase)

	if fractionalPart < rsg.DutyCycle {
		return rsg.Amplitude
	}
	return -rsg.Amplitude
}

// generateSawtooth генерирует пилообразный сигнал
func (rsg *ReferenceSignalGenerator) generateSawtooth(normalizedTime float64) float64 {
	// Фаза с учетом начальной фазы
	phase := normalizedTime + rsg.Phase/(2*math.Pi)
	fractionalPart := phase - math.Floor(phase)

	// Линейный рост от -Amplitude до Amplitude
	return rsg.Amplitude * (2*fractionalPart - 1)
}

// generateTriangle генерирует треугольный сигнал
func (rsg *ReferenceSignalGenerator) generateTriangle(normalizedTime float64) float64 {
	// Фаза с учетом начальной фазы
	phase := normalizedTime + rsg.Phase/(2*math.Pi)
	fractionalPart := phase - math.Floor(phase)

	// Треугольный сигнал
	if fractionalPart < 0.25 {
		return rsg.Amplitude * 4 * fractionalPart
	} else if fractionalPart < 0.75 {
		return rsg.Amplitude * (2 - 4*fractionalPart)
	} else {
		return rsg.Amplitude * (4*fractionalPart - 4)
	}
}

// validate проверяет корректность параметров
func (rsg *ReferenceSignalGenerator) validate() error {
	if rsg.Frequency <= 0 {
		return fmt.Errorf("частота должна быть положительной: %f", rsg.Frequency)
	}
	if rsg.SampleRate <= 0 {
		return fmt.Errorf("частота дискретизации должна быть положительной: %f", rsg.SampleRate)
	}
	if rsg.TotalTime <= 0 {
		return fmt.Errorf("длительность должна быть положительной: %f", rsg.TotalTime)
	}
	if rsg.Amplitude <= 0 {
		return fmt.Errorf("амплитуда должна быть положительной: %f", rsg.Amplitude)
	}
	if rsg.DutyCycle <= 0 || rsg.DutyCycle >= 1 {
		return fmt.Errorf("коэффициент заполнения должен быть в диапазоне (0, 1): %f", rsg.DutyCycle)
	}

	// Проверка критерия Найквиста
	if rsg.Frequency*2 >= rsg.SampleRate {
		return fmt.Errorf(
			"нарушен критерий Найквиста: частота сигнала (%f Гц) должна быть меньше половины частоты дискретизации (%f Гц)",
			rsg.Frequency, rsg.SampleRate/2,
		)
	}

	return nil
}

// Info возвращает информацию о настройках генератора
func (rsg *ReferenceSignalGenerator) Info() string {
	return fmt.Sprintf(
		"Тип сигнала: %s\nЧастота: %.1f Гц\nЧастота дискретизации: %.1f Гц\n"+
			"Длительность: %.1f с\nАмплитуда: %.1f\nНачальная фаза: %.2f рад\n"+
			"Коэффициент заполнения: %.1f%%\nКоличество отсчётов: %d\n"+
			"Период сигнала: %.4f с (%.1f отсчётов)",
		rsg.SignalType,
		rsg.Frequency,
		rsg.SampleRate,
		rsg.TotalTime,
		rsg.Amplitude,
		rsg.Phase,
		rsg.DutyCycle*100,
		int(math.Round(rsg.TotalTime*rsg.SampleRate)),
		1/rsg.Frequency,
		rsg.SampleRate/rsg.Frequency,
	)
}
