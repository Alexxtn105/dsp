package fft_alt

import (
	"fmt"
	"math"
	"math/cmplx"
)

const pi = math.Pi

// isPowerOfTwo проверяет, является ли число степенью двойки
func isPowerOfTwo(n int) bool {
	return n > 0 && (n&(n-1)) == 0
}

// FFT выполняет прямое преобразование Фурье (исправленная версия)
func fft(x []complex128) []complex128 {
	N := len(x)
	if N <= 1 {
		return x
	}

	// Разделяем на чётные и нечётные индексы
	even := make([]complex128, N/2)
	odd := make([]complex128, N/2)
	for i := 0; i < N/2; i++ {
		even[i] = x[2*i]
		odd[i] = x[2*i+1]
	}

	evenFFT := fft(even)
	oddFFT := fft(odd)

	result := make([]complex128, N)
	for k := 0; k < N/2; k++ {
		// Twiddle factor: W_N^k = e^(-2πik/N)
		twiddle := cmplx.Exp(complex(0, -2*pi*float64(k)/float64(N)))
		result[k] = evenFFT[k] + twiddle*oddFFT[k]
		result[k+N/2] = evenFFT[k] - twiddle*oddFFT[k]
	}

	return result
}

// WindowFunction определяет тип оконной функции
type WindowFunction int

const (
	Rectangular WindowFunction = iota
	Hann
	Hamming
	Blackman
)

// generateWindow создаёт коэффициенты оконной функции
func generateWindow(size int, windowType WindowFunction) []float64 {
	window := make([]float64, size)

	switch windowType {
	case Rectangular:
		for i := range window {
			window[i] = 1.0
		}
	case Hann:
		for i := range window {
			window[i] = 0.5 * (1.0 - math.Cos(2*pi*float64(i)/float64(size-1)))
		}
	case Hamming:
		for i := range window {
			window[i] = 0.54 - 0.46*math.Cos(2*pi*float64(i)/float64(size-1))
		}
	case Blackman:
		for i := range window {
			window[i] = 0.42 - 0.5*math.Cos(2*pi*float64(i)/float64(size-1)) +
				0.08*math.Cos(4*pi*float64(i)/float64(size-1))
		}
	}

	return window
}

// SlidingFFT представляет собой структуру скользящего окна для обновления спектра
type SlidingFFT struct {
	windowSize  int          // Размер окна БПФ (должен быть степенью 2)
	buffer      []float64    // Циркулярный буфер с отсчётами
	spectrum    []complex128 // Текущий спектр
	window      []float64    // Коэффициенты оконной функции
	writePos    int          // Позиция записи в циркулярном буфере
	initialized bool         // Флаг инициализации буфера
}

// NewSlidingFFT создаёт новый экземпляр скользящего БПФ
func NewSlidingFFT(windowSize int, windowType WindowFunction) (*SlidingFFT, error) {
	if !isPowerOfTwo(windowSize) {
		return nil, fmt.Errorf("размер окна должен быть степенью 2, получено: %d", windowSize)
	}

	return &SlidingFFT{
		windowSize:  windowSize,
		buffer:      make([]float64, windowSize),
		spectrum:    make([]complex128, windowSize),
		window:      generateWindow(windowSize, windowType),
		writePos:    0,
		initialized: false,
	}, nil
}

// Update добавляет новый отсчёт и пересчитывает спектр
func (sf *SlidingFFT) Update(newSample float64) {
	// Циркулярная запись в буфер
	sf.buffer[sf.writePos] = newSample
	sf.writePos = (sf.writePos + 1) % sf.windowSize

	// Отмечаем, что буфер заполнен после первого полного цикла
	if sf.writePos == 0 {
		sf.initialized = true
	}

	// Пересчитываем спектр только если буфер заполнен
	if sf.initialized {
		sf.computeSpectrum()
	}
}

// computeSpectrum вычисляет спектр из текущего состояния буфера
func (sf *SlidingFFT) computeSpectrum() {
	// Создаём временный массив с применением оконной функции
	// Данные берутся от writePos (самые старые) до writePos-1 (самые новые)
	complexInput := make([]complex128, sf.windowSize)

	for i := 0; i < sf.windowSize; i++ {
		// Индекс в циркулярном буфере (от самого старого к новому)
		bufferIdx := (sf.writePos + i) % sf.windowSize
		// Применяем оконную функцию
		windowedSample := sf.buffer[bufferIdx] * sf.window[i]
		complexInput[i] = complex(windowedSample, 0)
	}

	// Вычисляем БПФ
	sf.spectrum = fft(complexInput)
}

// GetSpectrum возвращает текущий спектр
func (sf *SlidingFFT) GetSpectrum() []complex128 {
	if !sf.initialized {
		return nil // Возвращаем nil, если буфер ещё не заполнен
	}
	return sf.spectrum
}

// GetMagnitudeSpectrum возвращает амплитудный спектр
func (sf *SlidingFFT) GetMagnitudeSpectrum() []float64 {
	if !sf.initialized {
		return nil
	}

	magnitude := make([]float64, sf.windowSize)
	for i, c := range sf.spectrum {
		magnitude[i] = cmplx.Abs(c)
	}
	return magnitude
}

// GetPowerSpectrum возвращает спектр мощности
func (sf *SlidingFFT) GetPowerSpectrum() []float64 {
	if !sf.initialized {
		return nil
	}

	power := make([]float64, sf.windowSize)
	for i, c := range sf.spectrum {
		mag := cmplx.Abs(c)
		power[i] = mag * mag
	}
	return power
}

// GetPowerSpectrumdB возвращает спектр мощности в дБ
func (sf *SlidingFFT) GetPowerSpectrumdB() []float64 {
	if !sf.initialized {
		return nil
	}

	power := sf.GetPowerSpectrum()
	powerDB := make([]float64, len(power))

	for i, p := range power {
		if p > 0 {
			powerDB[i] = 10 * math.Log10(p)
		} else {
			powerDB[i] = -200 // Минимальное значение для нулевой мощности
		}
	}

	return powerDB
}

// IsInitialized проверяет, заполнен ли буфер
func (sf *SlidingFFT) IsInitialized() bool {
	return sf.initialized
}
