package fft

import (
	//"fmt"
	"math"
	"math/cmplx"
)

const pi = math.Pi

// SlidingFFT реализует настоящий скользящий FFT с обновлением O(N)
type SlidingFFT struct {
	windowSize int          // Размер окна (должен быть степенью 2)
	spectrum   []complex128 // Текущий спектр
	buffer     []float64    // Кольцевой буфер
	pos        int          // Текущая позиция в кольцевом буфере
	twiddles   []complex128 // Предвычисленные twiddle factors
	cosTable   []float64    // Таблица косинусов для оптимизации
	sinTable   []float64    // Таблица синусов для оптимизации
	normFactor float64      // Коэффициент нормализации
}

// NewSlidingFFT создает новый скользящий FFT
func NewSlidingFFT(windowSize int) *SlidingFFT {
	// Проверяем, что размер - степень двойки
	if windowSize&(windowSize-1) != 0 {
		panic("windowSize must be a power of 2")
	}

	s := &SlidingFFT{
		windowSize: windowSize,
		spectrum:   make([]complex128, windowSize),
		buffer:     make([]float64, windowSize),
		twiddles:   make([]complex128, windowSize),
		cosTable:   make([]float64, windowSize),
		sinTable:   make([]float64, windowSize),
		normFactor: 1.0 / float64(windowSize),
		pos:        0,
	}

	// Предвычисляем twiddle factors
	for k := 0; k < windowSize; k++ {
		angle := -2 * pi * float64(k) / float64(windowSize)
		s.twiddles[k] = cmplx.Exp(complex(0, angle))
		s.cosTable[k] = math.Cos(angle)
		s.sinTable[k] = math.Sin(angle)
	}

	return s
}

// Initialize инициализирует FFT начальными данными
func (s *SlidingFFT) Initialize(initialSamples []float64) {
	if len(initialSamples) != s.windowSize {
		panic("initialSamples must have length equal to windowSize")
	}

	// Копируем начальные данные
	copy(s.buffer, initialSamples)
	s.pos = 0

	// Вычисляем начальный спектр через прямое FFT
	complexInput := make([]complex128, s.windowSize)
	for i := 0; i < s.windowSize; i++ {
		complexInput[i] = complex(s.buffer[i], 0)
	}
	s.spectrum = s.fft(complexInput)
}

// Update добавляет новый отсчет и обновляет спектр за O(N)
func (s *SlidingFFT) Update(newSample float64) {
	oldSample := s.buffer[s.pos]       // Самый старый отсчет, который выходит из окна
	s.buffer[s.pos] = newSample        // Заменяем его новым
	s.pos = (s.pos + 1) % s.windowSize // Перемещаем позицию

	// Обновляем спектр используя свойство сдвига Фурье
	delta := (newSample - oldSample) * s.normFactor

	// Оптимизированная версия обновления спектра
	for k := 0; k < s.windowSize; k++ {
		// X_k_new = e^(2πjk/N) * (X_k_old + delta)
		// Но так как мы используем forward transform с отрицательным знаком,
		// формула немного отличается

		// Более стабильная версия с использованием предвычисленных таблиц
		realPart := real(s.spectrum[k])
		imagPart := imag(s.spectrum[k])

		// Добавляем дельту (одинаково для всех k)
		realPart += delta

		// Поворачиваем фазовый множитель
		cosVal := s.cosTable[k]
		sinVal := s.sinTable[k]

		// Умножение комплексного числа на e^(j*theta) = cos(theta) + j*sin(theta)
		newReal := realPart*cosVal - imagPart*sinVal
		newImag := realPart*sinVal + imagPart*cosVal

		s.spectrum[k] = complex(newReal, newImag)
	}
}

// fft реализует стандартное БПФ для инициализации
func (s *SlidingFFT) fft(x []complex128) []complex128 {
	N := len(x)

	// Бит-реверс перестановка
	j := 0
	for i := 1; i < N; i++ {
		bit := N >> 1
		for j >= bit {
			j -= bit
			bit >>= 1
		}
		j += bit
		if i < j {
			x[i], x[j] = x[j], x[i]
		}
	}

	// Вычисления по бабочке
	for length := 2; length <= N; length <<= 1 {
		halfLen := length >> 1
		step := N / length

		for i := 0; i < N; i += length {
			k := 0
			for j := 0; j < halfLen; j++ {
				idx := k % N
				t := x[i+j+halfLen] * s.twiddles[idx]
				x[i+j+halfLen] = x[i+j] - t
				x[i+j] = x[i+j] + t
				k += step
			}
		}
	}
	return x
}

// GetSpectrum возвращает копию текущего спектра
func (s *SlidingFFT) GetSpectrum() []complex128 {
	spectrum := make([]complex128, s.windowSize)
	copy(spectrum, s.spectrum)
	return spectrum
}

// GetMagnitude возвращает амплитудный спектр
func (s *SlidingFFT) GetMagnitude() []float64 {
	magnitude := make([]float64, s.windowSize/2+1)
	for i := 0; i <= s.windowSize/2; i++ {
		magnitude[i] = cmplx.Abs(s.spectrum[i])
	}
	return magnitude
}

// GetPhase возвращает фазовый спектр
func (s *SlidingFFT) GetPhase() []float64 {
	phase := make([]float64, s.windowSize/2+1)
	for i := 0; i <= s.windowSize/2; i++ {
		phase[i] = math.Atan2(imag(s.spectrum[i]), real(s.spectrum[i]))
	}
	return phase
}

// AddWindow применяет оконную функцию к данным
func (s *SlidingFFT) AddWindow(windowType string) {
	var window []float64

	switch windowType {
	case "hann":
		window = s.hannWindow()
	case "hamming":
		window = s.hammingWindow()
	case "blackman":
		window = s.blackmanWindow()
	default:
		return
	}

	// Применяем окно к данным в буфере
	for i := 0; i < s.windowSize; i++ {
		s.buffer[i] *= window[i]
	}

	// Пересчитываем спектр
	s.recalculateSpectrum()
}

// Вспомогательные методы для оконных функций
func (s *SlidingFFT) hannWindow() []float64 {
	window := make([]float64, s.windowSize)
	for i := 0; i < s.windowSize; i++ {
		window[i] = 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(s.windowSize-1)))
	}
	return window
}

func (s *SlidingFFT) hammingWindow() []float64 {
	window := make([]float64, s.windowSize)
	alpha, beta := 0.54, 0.46
	for i := 0; i < s.windowSize; i++ {
		window[i] = alpha - beta*math.Cos(2*math.Pi*float64(i)/float64(s.windowSize-1))
	}
	return window
}

func (s *SlidingFFT) blackmanWindow() []float64 {
	window := make([]float64, s.windowSize)
	a0, a1, a2 := 0.42, 0.5, 0.08
	for i := 0; i < s.windowSize; i++ {
		x := 2 * math.Pi * float64(i) / float64(s.windowSize-1)
		window[i] = a0 - a1*math.Cos(x) + a2*math.Cos(2*x)
	}
	return window
}

// recalculateSpectrum полностью пересчитывает спектр (используется после применения окна)
func (s *SlidingFFT) recalculateSpectrum() {
	complexInput := make([]complex128, s.windowSize)
	for i := 0; i < s.windowSize; i++ {
		idx := (s.pos + i) % s.windowSize
		complexInput[i] = complex(s.buffer[idx], 0)
	}
	s.spectrum = s.fft(complexInput)
}
