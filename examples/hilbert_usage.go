package main

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/Alexxtn105/dsp/hilbert"
)

func main() {
	// Создаем преобразователь Гильберта
	sampleRate := 48000.0 // 48 кГц
	order := 63           // Порядок фильтра

	ht := hilbert.NewHilbertTransform(sampleRate, order)

	// Генерируем тестовый синусоидальный сигнал
	frequency := 1000.0 // 1 кГц
	numSamples := 1000

	for i := 0; i < numSamples; i++ {
		t := float64(i) / sampleRate
		input := math.Sin(2.0 * math.Pi * frequency * t)

		// Обрабатываем отсчет
		output := ht.Tick(input)

		// После начальной переходной характеристики фильтра
		if i > ht.GetGroupDelay()*2 {
			magnitude := cmplx.Abs(output)
			phase := cmplx.Phase(output)

			fmt.Printf("Sample %d: |z| = %.4f, ∠z = %.4f rad\n",
				i, magnitude, phase)
		}
	}
}
