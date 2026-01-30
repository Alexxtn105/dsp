package filters

import (
	"math"
	"testing"
)

// TestFIRFilterBasic проверяет базовую функциональность фильтра
func TestFIRFilterBasic(t *testing.T) {
	coeffs := []float64{1.0, 2.0, 3.0}
	filter := NewFIRFilter(coeffs)

	// Импульсный отклик
	outputs := []float64{
		filter.Tick(1.0), // 1*1 = 1
		filter.Tick(0.0), // 1*0 + 2*1 = 2
		filter.Tick(0.0), // 1*0 + 2*0 + 3*1 = 3
		filter.Tick(0.0), // 1*0 + 2*0 + 3*0 = 0
	}

	expected := []float64{1.0, 2.0, 3.0, 0.0}
	for i, val := range outputs {
		if math.Abs(val-expected[i]) > 1e-10 {
			t.Errorf("Tick %d: ожидалось %f, получено %f", i, expected[i], val)
		}
	}
}

// TestFIRFilterImpulseResponse проверяет импульсную характеристику
func TestFIRFilterImpulseResponse(t *testing.T) {
	coeffs := []float64{0.5, -0.2, 0.1, 0.3}
	filter := NewFIRFilter(coeffs)

	// Единичный импульс
	filter.Tick(1.0)

	// Должны получить коэффициенты в обратном порядке
	outputs := make([]float64, len(coeffs))
	for i := 0; i < len(coeffs); i++ {
		if i == 0 {
			outputs[i] = filter.Tick(0.0)
		} else {
			outputs[i] = filter.Tick(0.0)
		}
	}

	// Проверяем импульсную характеристику
	for i, coeff := range coeffs {
		if i > 0 { // Первый коэффициент уже был применен
			if math.Abs(outputs[i-1]-coeff) > 1e-10 {
				t.Errorf("Коэффициент %d: ожидалось %f, получено %f",
					i, coeff, outputs[i-1])
			}
		}
	}
}

// TestAllOnesCoefficients проверяет фильтр со всеми коэффициентами = 1
func TestAllOnesCoefficients(t *testing.T) {
	n := 5
	coeffs := make([]float64, n)
	for i := range coeffs {
		coeffs[i] = 1.0
	}

	filter := NewFIRFilter(coeffs)

	// Единичный импульс
	filter.Tick(1.0)

	// После импульса должны получить сумму предыдущих значений
	outputs := make([]float64, n+2)
	outputs[0] = 1.0 // Первый выход уже получен

	for i := 1; i < n+2; i++ {
		if i < n {
			outputs[i] = filter.Tick(0.0)
		} else {
			outputs[i] = filter.Tick(0.0)
		}
	}

	// Для фильтра [1,1,1,1,1] с входом [1,0,0,0,0,0...]
	// Выход: [1,1,1,1,1,0,0...]
	expected := []float64{1.0, 1.0, 1.0, 1.0, 1.0, 0.0, 0.0}

	for i, val := range outputs {
		if math.Abs(val-expected[i]) > 1e-10 {
			t.Errorf("Позиция %d: ожидалось %f, получено %f",
				i, expected[i], val)
		}
	}

	// Проверяем сумму коэффициентов
	var sum float64
	for _, c := range coeffs {
		sum += c
	}
	if math.Abs(sum-float64(n)) > 1e-10 {
		t.Errorf("Сумма коэффициентов: ожидалось %f, получено %f",
			float64(n), sum)
	}
}

// TestSingleCoefficientFilter проверяет фильтр с одним коэффициентом
func TestSingleCoefficientFilter(t *testing.T) {
	// Нейтральный фильтр без задержки
	filter := NewFIRFilter([]float64{1.0})

	testSignal := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	for i, input := range testSignal {
		output := filter.Tick(input)
		if math.Abs(output-input) > 1e-10 {
			t.Errorf("Тик %d: ожидалось %f, получено %f", i, input, output)
		}
	}

	// Фильтр с усилением
	gainFilter := NewFIRFilter([]float64{2.5})
	for i, input := range testSignal {
		output := gainFilter.Tick(input)
		expected := input * 2.5
		if math.Abs(output-expected) > 1e-10 {
			t.Errorf("Тик %d с усилением: ожидалось %f, получено %f",
				i, expected, output)
		}
	}
}

// TestMovingAverageFilter проверяет фильтр скользящего среднего
func TestMovingAverageFilter(t *testing.T) {
	n := 4
	coeffs := make([]float64, n)
	for i := range coeffs {
		coeffs[i] = 1.0 / float64(n) // Нормализованный
	}

	filter := NewFIRFilter(coeffs)

	// Постоянный сигнал
	constantValue := 3.14
	for i := 0; i < 10; i++ {
		output := filter.Tick(constantValue)
		if i < n-1 {
			// Пока буфер не заполнен
			expected := constantValue * float64(i+1) / float64(n)
			if math.Abs(output-expected) > 1e-10 {
				t.Errorf("Заполнение буфера, тик %d: ожидалось %f, получено %f",
					i, expected, output)
			}
		} else {
			// После заполнения буфера
			if math.Abs(output-constantValue) > 1e-10 {
				t.Errorf("Установившийся режим, тик %d: ожидалось %f, получено %f",
					i, constantValue, output)
			}
		}
	}

	// Проверяем сумму коэффициентов = 1
	var sum float64
	for _, c := range coeffs {
		sum += c
	}
	if math.Abs(sum-1.0) > 1e-10 {
		t.Errorf("Сумма коэффициентов скользящего среднего: ожидалось 1.0, получено %f", sum)
	}
}

// TestNeutralFilterDelay проверяет нейтральный фильтр с задержкой
func TestNeutralFilterDelay(t *testing.T) {
	// Задержка на 3 отсчета: [0, 0, 0, 1]
	coeffs := []float64{0, 0, 0, 1}
	filter := NewFIRFilter(coeffs)

	testSignal := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0}

	for i, input := range testSignal {
		output := filter.Tick(input)
		if i < 3 {
			// Первые 3 отсчета - задержка
			if math.Abs(output) > 1e-10 {
				t.Errorf("Задержка, тик %d: ожидалось 0, получено %f", i, output)
			}
		} else {
			// После задержки получаем исходный сигнал с задержкой
			expected := testSignal[i-3]
			if math.Abs(output-expected) > 1e-10 {
				t.Errorf("После задержки, тик %d: ожидалось %f, получено %f",
					i, expected, output)
			}
		}
	}
}

// TestZeroCoefficients проверяет фильтр с нулевыми коэффициентами
func TestZeroCoefficients(t *testing.T) {
	coeffs := []float64{0, 0, 0, 0}
	filter := NewFIRFilter(coeffs)

	// Любой сигнал должен давать 0 на выходе
	for i := 0; i < 10; i++ {
		output := filter.Tick(float64(i))
		if math.Abs(output) > 1e-10 {
			t.Errorf("Тик %d: ожидалось 0, получено %f", i, output)
		}
	}
}

// TestEmptyCoefficients проверяет обработку пустых коэффициентов
func TestEmptyCoefficients(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Ожидалась паника при пустых коэффициентах")
		}
	}()

	_ = NewFIRFilter([]float64{})
}

// TestResetFunctionality проверяет сброс фильтра
func TestResetFunctionality(t *testing.T) {
	coeffs := []float64{1, 2, 3}
	filter := NewFIRFilter(coeffs)

	// Обрабатываем некоторый сигнал
	for i := 0; i < 5; i++ {
		filter.Tick(float64(i))
	}

	// Сбрасываем
	filter.Reset()

	// После сброса фильтр должен вести себя как новый
	output := filter.Tick(1.0)
	expected := 1.0 // 1 * 1 = 1
	if math.Abs(output-expected) > 1e-10 {
		t.Errorf("После Reset: ожидалось %f, получено %f", expected, output)
	}
}

// TestConsecutiveProcessing проверяет последовательную обработку
func TestConsecutiveProcessing(t *testing.T) {
	coeffs := []float64{0.5, -0.2}
	filter := NewFIRFilter(coeffs)

	// Длинная последовательность
	inputs := []float64{1, -1, 2, -2, 3, -3}
	expected := []float64{
		0.5,  // 0.5*1 + (-0.2)*0 = 0.5
		-0.7, // 0.5*(-1) + (-0.2)*1 = -0.5 - 0.2 = -0.7
		1.2,  // 0.5*2 + (-0.2)*(-1) = 1.0 + 0.2 = 1.2
		-1.4, // 0.5*(-2) + (-0.2)*2 = -1.0 - 0.4 = -1.4
		1.9,  // 0.5*3 + (-0.2)*(-2) = 1.5 + 0.4 = 1.9
		-2.1, // 0.5*(-3) + (-0.2)*3 = -1.5 - 0.6 = -2.1
	}

	for i, input := range inputs {
		output := filter.Tick(input)
		if math.Abs(output-expected[i]) > 1e-10 {
			t.Errorf("Тик %d: ожидалось %f, получено %f",
				i, expected[i], output)
		}
	}
}

// BenchmarkFIRFilterTick тестирует производительность
func BenchmarkFIRFilterTick(b *testing.B) {
	// Фильтр с 64 коэффициентами
	coeffs := make([]float64, 64)
	for i := range coeffs {
		coeffs[i] = 1.0 / 64.0 // Скользящее среднее
	}

	filter := NewFIRFilter(coeffs)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.Tick(float64(i))
	}
}
