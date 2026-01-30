package main

import (
	"fmt"
	"math"

	"github.com/Alexxtn105/dsp/fft"
)

// Пример использования
func main() {
	// Создаем скользящий FFT с окном 1024 отсчета
	windowSize := 1024
	fftSliding := fft.NewSlidingFFT(windowSize)

	// Инициализируем начальными данными (например, нулями)
	initialData := make([]float64, windowSize)
	for i := range initialData {
		initialData[i] = math.Sin(2 * math.Pi * 50 * float64(i) / 1000.0) // 50 Гц при Fs=1000 Гц
	}

	fftSliding.Initialize(initialData)

	// Применяем оконную функцию для уменьшения спектральных протечек
	fftSliding.AddWindow("hann")

	// Обновляем FFT с новыми отсчетами
	for i := 0; i < 100; i++ {
		// Генерируем новый отсчет (например, 50 Гц синусоида с небольшим шумом)
		newSample := math.Sin(2*math.Pi*50*float64(windowSize+i)/1000.0) + 0.1*math.Sin(2*math.Pi*120*float64(windowSize+i)/1000.0)
		fftSliding.Update(newSample)

		// Получаем текущий спектр
		//spectrum := fftSliding.GetSpectrum()
		magnitude := fftSliding.GetMagnitude()

		// Находим пик в спектре
		maxFreq := 0
		maxMag := 0.0
		for freq, mag := range magnitude {
			if mag > maxMag {
				maxMag = mag
				maxFreq = freq
			}
		}

		fmt.Printf("Update %d: Peak at frequency bin %d (magnitude: %.4f)\n",
			i+1, maxFreq, maxMag)

		// Можно также получить фазовый спектр
		phase := fftSliding.GetPhase()
		if maxFreq < len(phase) {
			fmt.Printf("  Phase at peak: %.4f rad\n", phase[maxFreq])
		}
	}

	fmt.Println("\nСкользящий FFT успешно обновлен 100 раз!")

	// Демонстрация работы с несколькими частотами
	fmt.Println("\nДемонстрация с двумя частотами:")
	fft2 := fft.NewSlidingFFT(256)

	// Тест с двумя частотами
	testData := make([]float64, 256)
	for i := range testData {
		testData[i] = math.Sin(2*math.Pi*20*float64(i)/256.0) +
			0.5*math.Sin(2*math.Pi*60*float64(i)/256.0)
	}

	fft2.Initialize(testData)
	fft2.AddWindow("hamming")

	mag := fft2.GetMagnitude()
	fmt.Println("Основные частотные компоненты:")
	for freq, magnitude := range mag {
		if magnitude > 0.1 { // Порог для значимых компонент
			actualFreq := float64(freq) * 1000.0 / 256.0 // Предполагая Fs=1000 Гц
			fmt.Printf("  %6.1f Гц: магнитуда = %.4f\n", actualFreq, magnitude)
		}
	}
}
