package main

import (
	"fmt"
	"math"

	"github.com/Alexxtn105/dsp/detectors"
	"github.com/Alexxtn105/dsp/generators"
)

func main() {
	// Генерация тестового сигнала
	testSignal := GenerateTestComplexSignal(1000, 8000, 0.1)

	// Создание частотного детектора
	detector := detectors.NewFrequencyDetector(8000)

	// Обработка сигнала
	frequencies := detector.ProcessBlock(testSignal)

	// Вывод первых 10 значений частоты
	for i := 0; i < 10 && i < len(frequencies); i++ {
		fmt.Printf("Sample %d: Frequency = %.2f Hz\n", i, frequencies[i])
	}

	// Альтернативный вариант с PLL
	pllDetector := detectors.NewPLLFrequencyDetector(8000, 100) // Полоса 100 Гц
	pllFrequencies := make([]float64, len(testSignal))
	for i, sig := range testSignal {
		pllFrequencies[i] = pllDetector.DetectFrequencyPLL(sig)
	}
}

// GenerateTestComplexSignal Helper функция для генерации тестового комплексного сигнала
func GenerateTestComplexSignal(freq, sampleRate, duration float64) []complex128 {
	rsg := generators.NewReferenceSignalGenerator()
	rsg.Frequency = freq
	rsg.SampleRate = sampleRate
	rsg.TotalTime = duration
	rsg.SignalType = generators.Cosine

	// Генерируем I и Q компоненты
	iSignal, _ := rsg.Generate()

	// Для теста создаем квадратурный сигнал со сдвигом на 90 градусов
	rsg.Phase = math.Pi / 2
	qSignal, _ := rsg.Generate()

	// Формируем комплексный сигнал
	complexSignal := make([]complex128, len(iSignal))
	for i := range iSignal {
		complexSignal[i] = complex(iSignal[i], qSignal[i])
	}

	return complexSignal
}
