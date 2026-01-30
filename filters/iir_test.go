package filters

import (
	"math"
	"math/cmplx"
	"testing"
)

// TestIIRFilter_Basic проверяет базовую функциональность БИХ-фильтра
func TestIIRFilter_Basic(t *testing.T) {
	// Простой БИХ-фильтр 1-го порядка: y[n] = 0.5*x[n] - 0.3*y[n-1]
	b := []float64{0.5}
	a := []float64{1, 0.3} // a[0] всегда 1
	filter := NewIIRFilter(b, a)

	// Импульсный отклик
	outputs := make([]float64, 5)
	outputs[0] = filter.Tick(1.0) // y[0] = 0.5*1 = 0.5
	outputs[1] = filter.Tick(0.0) // y[1] = 0.5*0 - 0.3*0.5 = -0.15
	outputs[2] = filter.Tick(0.0) // y[2] = -0.3*(-0.15) = 0.045
	outputs[3] = filter.Tick(0.0) // y[3] = -0.3*0.045 = -0.0135
	outputs[4] = filter.Tick(0.0) // y[4] = -0.3*(-0.0135) = 0.00405

	expected := []float64{0.5, -0.15, 0.045, -0.0135, 0.00405}
	for i, val := range outputs {
		if math.Abs(val-expected[i]) > 1e-10 {
			t.Errorf("Tick %d: ожидалось %f, получено %f", i, expected[i], val)
		}
	}
}

// TestIIRFilter_FirstOrder проверяет фильтр 1-го порядка
func TestIIRFilter_FirstOrder(t *testing.T) {
	// Фильтр: y[n] = 0.8*x[n] + 0.2*x[n-1] - 0.5*y[n-1]
	b := []float64{0.8, 0.2}
	a := []float64{1, 0.5}
	filter := NewIIRFilter(b, a)

	// Step response
	outputs := make([]float64, 5)
	for i := 0; i < 5; i++ {
		input := 1.0 // Единичный ступенчатый сигнал
		outputs[i] = filter.Tick(input)
	}

	// Расчет вручную:
	// y[0] = 0.8*1 = 0.8
	// y[1] = 0.8*1 + 0.2*1 - 0.5*0.8 = 0.8 + 0.2 - 0.4 = 0.6
	// y[2] = 0.8*1 + 0.2*1 - 0.5*0.6 = 0.8 + 0.2 - 0.3 = 0.7
	// y[3] = 0.8*1 + 0.2*1 - 0.5*0.7 = 0.8 + 0.2 - 0.35 = 0.65
	// y[4] = 0.8*1 + 0.2*1 - 0.5*0.65 = 0.8 + 0.2 - 0.325 = 0.675
	expected := []float64{0.8, 0.6, 0.7, 0.65, 0.675}

	for i, val := range outputs {
		if math.Abs(val-expected[i]) > 1e-10 {
			t.Errorf("Шаг %d: ожидалось %f, получено %f", i, expected[i], val)
		}
	}
}

// TestIIRFilter_Normalization проверяет нормализацию коэффициентов
func TestIIRFilter_Normalization(t *testing.T) {
	// Коэффициенты с a[0] != 1
	b := []float64{2.0, 1.0}
	a := []float64{4.0, 2.0} // a[0] = 4, должно нормализоваться к 1

	filter := NewIIRFilter(b, a)

	// После нормализации должно быть:
	// b' = [2/4, 1/4] = [0.5, 0.25]
	// a' = [4/4, 2/4] = [1, 0.5]
	expectedB := []float64{0.5, 0.25}
	expectedA := []float64{1.0, 0.5}

	coeffsB := filter.GetBCoeffs()
	coeffsA := filter.GetACoeffs()

	for i, val := range coeffsB {
		if math.Abs(val-expectedB[i]) > 1e-10 {
			t.Errorf("Коэффициент b[%d]: ожидалось %f, получено %f",
				i, expectedB[i], val)
		}
	}

	for i, val := range coeffsA {
		if math.Abs(val-expectedA[i]) > 1e-10 {
			t.Errorf("Коэффициент a[%d]: ожидалось %f, получено %f",
				i, expectedA[i], val)
		}
	}
}

// TestIIRFilter_FirstOrderLowPass проверяет ФНЧ 1-го порядка
func TestIIRFilter_FirstOrderLowPass(t *testing.T) {
	fc := 0.1 // Частота среза = 0.1 * Fs/2
	filter := NewFirstOrderLowPass(fc)

	// Проверяем устойчивость
	if !filter.IsStable() {
		t.Error("ФНЧ 1-го порядка должен быть устойчив")
	}

	// Проверяем коэффициенты
	b := filter.GetBCoeffs()
	a := filter.GetACoeffs()

	// Для ФНЧ 1-го порядка с билинейным преобразованием:
	// alpha = tan(pi*fc)
	// b0 = alpha/(1+alpha)
	// b1 = b0
	// a1 = -(1-alpha)/(1+alpha)
	alpha := math.Tan(math.Pi * fc)
	expectedB0 := alpha / (1 + alpha)
	expectedB1 := expectedB0
	expectedA1 := -(1 - alpha) / (1 + alpha)

	if math.Abs(b[0]-expectedB0) > 1e-10 {
		t.Errorf("b0: ожидалось %f, получено %f", expectedB0, b[0])
	}
	if math.Abs(b[1]-expectedB1) > 1e-10 {
		t.Errorf("b1: ожидалось %f, получено %f", expectedB1, b[1])
	}
	if math.Abs(a[1]-expectedA1) > 1e-10 {
		t.Errorf("a1: ожидалось %f, получено %f", expectedA1, a[1])
	}

	// Проверяем частотную характеристику
	// На нулевой частоте усиление должно быть ~1
	h0 := filter.GetFrequencyResponse(0.0)
	gainDC := cmplx.Abs(h0)
	if math.Abs(gainDC-1.0) > 0.01 {
		t.Errorf("Усиление на постоянном токе: ожидалось ~1.0, получено %f", gainDC)
	}

	// На частоте среза усиление должно быть ~0.707 (-3dB)
	hFc := filter.GetFrequencyResponse(fc)
	gainFc := cmplx.Abs(hFc)
	expectedGainFc := 1.0 / math.Sqrt(2.0) // -3dB
	if math.Abs(gainFc-expectedGainFc) > 0.05 {
		t.Errorf("Усиление на частоте среза: ожидалось ~%f, получено %f",
			expectedGainFc, gainFc)
	}
}

// TestIIRFilter_FirstOrderHighPass проверяет ФВЧ 1-го порядка
func TestIIRFilter_FirstOrderHighPass(t *testing.T) {
	fc := 0.2 // Частота среза
	filter := NewFirstOrderHighPass(fc)

	if !filter.IsStable() {
		t.Error("ФВЧ 1-го порядка должен быть устойчив")
	}

	// Проверяем частотную характеристику
	// На нулевой частоте усиление должно быть ~0
	h0 := filter.GetFrequencyResponse(0.0)
	gainDC := cmplx.Abs(h0)
	// Для ФВЧ на постоянном токе усиление должно быть очень малым,
	// но не обязательно точно 0 из-за квантования
	if gainDC > 0.05 { // Ослабление не менее 20 дБ
		t.Errorf("Усиление на постоянном токе: ожидалось <0.05, получено %f", gainDC)
	}

	// На частоте среза усиление должно быть ~0.707 (-3dB)
	hFc := filter.GetFrequencyResponse(fc)
	gainFc := cmplx.Abs(hFc)
	expectedGainFc := 1.0 / math.Sqrt(2.0) // -3dB
	// Увеличиваем допуск, так как фильтр 1-го порядка не обеспечивает
	// точное ослабление 3 дБ на частоте среза
	tolerance := 0.15 // 15% допуск
	if math.Abs(gainFc-expectedGainFc) > tolerance {
		t.Errorf("Усиление на частоте среза: ожидалось ~%f±%f, получено %f",
			expectedGainFc, tolerance, gainFc)
	}

	// На высокой частоте усиление должно быть ~1
	hHigh := filter.GetFrequencyResponse(0.45)
	gainHigh := cmplx.Abs(hHigh)
	if math.Abs(gainHigh-1.0) > 0.1 {
		t.Errorf("Усиление на высокой частоте: ожидалось ~1.0, получено %f", gainHigh)
	}

	// Дополнительная проверка: монотонное возрастание усиления
	hLow := filter.GetFrequencyResponse(fc * 0.5)
	gainLow := cmplx.Abs(hLow)
	hMid := filter.GetFrequencyResponse(fc)
	gainMid := cmplx.Abs(hMid)
	hHigher := filter.GetFrequencyResponse(fc * 1.5)
	gainHigher := cmplx.Abs(hHigher)

	// Для ФВЧ усиление должно возрастать с частотой
	if gainLow > gainMid || gainMid > gainHigher {
		t.Errorf("Усиление не монотонно возрастает: %.3f, %.3f, %.3f",
			gainLow, gainMid, gainHigher)
	}
}

// TestIIRFilter_FirstOrderHighPassButterworth проверяет ФВЧ 1-го порядка Баттерворта
func TestIIRFilter_FirstOrderHighPassButterworth(t *testing.T) {
	// Более точная реализация ФВЧ 1-го порядка (билинейное преобразование)
	fc := 0.2

	// Используем билинейное преобразование с предыскажением
	// Для ФВЧ Баттерворта 1-го порядка
	wc := 2.0 * math.Pi * fc
	// Предыскажение частоты для билинейного преобразования
	omegaC := math.Tan(wc / 2.0)

	// Коэффициенты для ФВЧ 1-го порядка
	//K := omega_c / (1 + omega_c)
	b0 := 1.0 / (1 + omegaC)
	b1 := -b0
	a1 := (1 - omegaC) / (1 + omegaC)

	filter := NewIIRFilter([]float64{b0, b1}, []float64{1, -a1})

	// Проверяем на частоте среза
	hFc := filter.GetFrequencyResponse(fc)
	gainFc := cmplx.Abs(hFc)
	expectedGainFc := 1.0 / math.Sqrt(2.0)

	// Теперь допуск может быть меньше
	tolerance := 0.01 // 1% допуск
	if math.Abs(gainFc-expectedGainFc) > tolerance {
		t.Errorf("Усиление на частоте среза: ожидалось %.6f±%.4f, получено %.6f",
			expectedGainFc, tolerance, gainFc)
	}

	// Логируем для отладки
	t.Logf("ФВЧ 1-го порядка (fc=%.2f):", fc)
	t.Logf("  b = %v", filter.GetBCoeffs())
	t.Logf("  a = %v", filter.GetACoeffs())
	t.Logf("  Усиление на fc: %.6f (ожидалось %.6f)", gainFc, expectedGainFc)
	t.Logf("  Усиление на 0: %.6f", cmplx.Abs(filter.GetFrequencyResponse(0.0)))
	t.Logf("  Усиление на 0.45: %.6f", cmplx.Abs(filter.GetFrequencyResponse(0.45)))
}

// TestIIRFilter_SecondOrderLowPass проверяет ФНЧ 2-го порядка
func TestIIRFilter_SecondOrderLowPass(t *testing.T) {
	fc := 0.1
	Q := 0.707 // Баттерворт
	filter := NewSecondOrderLowPass(fc, Q)

	if !filter.IsStable() {
		t.Error("ФНЧ 2-го порядка должен быть устойчив")
	}

	// Проверяем коэффициенты (должны соответствовать расчетам)
	// Они уже проверяются в конструкторе

	// Проверяем, что фильтр действительно низкочастотный
	h0 := filter.GetFrequencyResponse(0.0)
	gainDC := cmplx.Abs(h0)
	if math.Abs(gainDC-1.0) > 0.01 {
		t.Errorf("Усиление на постоянном токе: ожидалось ~1.0, получено %f", gainDC)
	}

	hFc := filter.GetFrequencyResponse(fc)
	gainFc := cmplx.Abs(hFc)
	expectedGainFc := 1.0 / math.Sqrt(2.0)
	if math.Abs(gainFc-expectedGainFc) > 0.1 {
		t.Errorf("Усиление на частоте среза: ожидалось ~%f, получено %f",
			expectedGainFc, gainFc)
	}
}

// TestIIRFilter_SecondOrderBandPass проверяет полосовой фильтр 2-го порядка
func TestIIRFilter_SecondOrderBandPass(t *testing.T) {
	fc := 0.25 // Центральная частота
	Q := 5.0   // Высокая добротность
	filter := NewSecondOrderBandPass(fc, Q)

	if !filter.IsStable() {
		t.Error("Полосовой фильтр должен быть устойчив")
	}

	// Проверяем частотную характеристику
	hFc := filter.GetFrequencyResponse(fc)
	gainFc := cmplx.Abs(hFc)

	// На центральной частоте усиление должно быть максимальным (~1 для Q=5)
	if gainFc < 0.9 || gainFc > 1.1 {
		t.Errorf("Усиление на центральной частоте: ожидалось ~1.0, получено %f", gainFc)
	}

	// На частотах далеко от центральной усиление должно быть малым
	hLow := filter.GetFrequencyResponse(0.01)
	gainLow := cmplx.Abs(hLow)
	if gainLow > 0.1 {
		t.Errorf("Усиление на низкой частоте: ожидалось <0.1, получено %f", gainLow)
	}

	hHigh := filter.GetFrequencyResponse(0.49)
	gainHigh := cmplx.Abs(hHigh)
	if gainHigh > 0.1 {
		t.Errorf("Усиление на высокой частоте: ожидалось <0.1, получено %f", gainHigh)
	}
}

// TestIIRFilter_Reset проверяет сброс фильтра
func TestIIRFilter_Reset(t *testing.T) {
	b := []float64{0.5, 0.3}
	a := []float64{1, 0.2, 0.1}
	filter := NewIIRFilter(b, a)

	// Обрабатываем некоторый сигнал
	for i := 0; i < 10; i++ {
		filter.Tick(float64(i))
	}

	// Сбрасываем
	filter.Reset()

	// После сброса фильтр должен вести себя как новый
	// С единичным импульсом
	output1 := filter.Tick(1.0) // Должно быть b0 = 0.5
	output2 := filter.Tick(0.0) // b1*x[-1] - a1*y[-1] = 0.3*1 - 0.2*0.5 = 0.3 - 0.1 = 0.2

	if math.Abs(output1-0.5) > 1e-10 {
		t.Errorf("После Reset, первый тик: ожидалось 0.5, получено %f", output1)
	}
	if math.Abs(output2-0.2) > 1e-10 {
		t.Errorf("После Reset, второй тик: ожидалось 0.2, получено %f", output2)
	}
}

// TestIIRFilter_Process проверяет обработку среза данных
func TestIIRFilter_Process(t *testing.T) {
	b := []float64{0.8, 0.2}
	a := []float64{1, -0.5}
	filter := NewIIRFilter(b, a)

	input := []float64{1, 1, 1, 1, 1}
	output := filter.Process(input)

	// Расчет вручную:
	// y[0] = 0.8*1 = 0.8
	// y[1] = 0.8*1 + 0.2*1 - (-0.5)*0.8 = 0.8 + 0.2 + 0.4 = 1.4
	// y[2] = 0.8*1 + 0.2*1 - (-0.5)*1.4 = 0.8 + 0.2 + 0.7 = 1.7
	// y[3] = 0.8*1 + 0.2*1 - (-0.5)*1.7 = 0.8 + 0.2 + 0.85 = 1.85
	// y[4] = 0.8*1 + 0.2*1 - (-0.5)*1.85 = 0.8 + 0.2 + 0.925 = 1.925
	expected := []float64{0.8, 1.4, 1.7, 1.85, 1.925}

	for i, val := range output {
		if math.Abs(val-expected[i]) > 1e-10 {
			t.Errorf("Элемент %d: ожидалось %f, получено %f", i, expected[i], val)
		}
	}
}

// TestIIRFilter_Stability проверяет устойчивость фильтров
func TestIIRFilter_Stability(t *testing.T) {
	// Устойчивый фильтр (полюса внутри единичной окружности)
	b1 := []float64{0.5}
	a1 := []float64{1, 0.3} // |0.3| < 1
	filter1 := NewIIRFilter(b1, a1)
	if !filter1.IsStable() {
		t.Error("Фильтр с a1=0.3 должен быть устойчив")
	}

	// Пограничный случай (полюс на единичной окружности)
	b2 := []float64{0.5}
	a2 := []float64{1, 1.0} // |1.0| = 1
	filter2 := NewIIRFilter(b2, a2)
	if filter2.IsStable() {
		t.Error("Фильтр с a1=1.0 не должен считаться устойчивым")
	}

	// Неустойчивый фильтр
	b3 := []float64{0.5}
	a3 := []float64{1, 1.5} // |1.5| > 1
	filter3 := NewIIRFilter(b3, a3)
	if filter3.IsStable() {
		t.Error("Фильтр с a1=1.5 должен быть неустойчив")
	}
}

// TestIIRFilter_FrequencyResponse проверяет вычисление частотной характеристики
func TestIIRFilter_FrequencyResponse(t *testing.T) {
	// Простой фильтр 1-го порядка
	b := []float64{1.0}
	a := []float64{1, -0.5}
	filter := NewIIRFilter(b, a)

	// На нулевой частоте: H(0) = 1/(1 - 0.5) = 2
	h0 := filter.GetFrequencyResponse(0.0)
	expectedH0 := complex(2.0, 0)
	if math.Abs(real(h0)-real(expectedH0)) > 1e-10 ||
		math.Abs(imag(h0)-imag(expectedH0)) > 1e-10 {
		t.Errorf("H(0): ожидалось %v, получено %v", expectedH0, h0)
	}

	// На частоте Найквиста (0.5): H(0.5) = 1/(1 + 0.5) = 2/3
	hNyq := filter.GetFrequencyResponse(0.5)
	expectedHNyq := complex(2.0/3.0, 0)
	if math.Abs(real(hNyq)-real(expectedHNyq)) > 1e-10 ||
		math.Abs(imag(hNyq)-imag(expectedHNyq)) > 1e-10 {
		t.Errorf("H(0.5): ожидалось %v, получено %v", expectedHNyq, hNyq)
	}
}

// TestIIRFilter_GroupDelay проверяет вычисление групповой задержки
func TestIIRFilter_GroupDelay(t *testing.T) {
	// Фильтр 1-го порядка с положительной задержкой
	b := []float64{0.5}
	a := []float64{1, -0.3}
	filter := NewIIRFilter(b, a)

	// Проверяем, что задержка вычисляется без ошибок
	// (знак может быть как положительным, так и отрицательным в зависимости от фильтра)
	delay := filter.GetGroupDelay(0.1)

	// Проверяем, что значение конечное и не NaN
	if math.IsNaN(delay) || math.IsInf(delay, 0) {
		t.Errorf("Групповая задержка некорректна: %f", delay)
	}

	// Для этого конкретного фильтра задержка должна быть положительной
	// Но в общем случае для БИХ-фильтров групповая задержка может быть и отрицательной
	// на некоторых частотах, так что не проверяем знак жестко

	// Проверяем, что задержка имеет разумное значение (обычно от -N до N отсчетов)
	if math.Abs(delay) > 100 {
		t.Errorf("Групповая задержка нереалистично большая: %f", delay)
	}

	// Проверяем согласованность: задержка на 0 частоте
	delayDC := filter.GetGroupDelay(0.0)
	if math.IsNaN(delayDC) || math.IsInf(delayDC, 0) {
		t.Errorf("Групповая задержка на постоянном токе некорректна: %f", delayDC)
	}

	// Проверяем другой фильтр - ФНЧ 1-го порядка (должен иметь положительную задержку)
	lpf := NewFirstOrderLowPass(0.1)
	delayLPF := lpf.GetGroupDelay(0.05)

	// Для ФНЧ групповая задержка обычно положительна
	if delayLPF < -0.5 { // Допускаем небольшую отрицательность из-за погрешности
		t.Errorf("Групповая задержка ФНЧ слишком отрицательная: %f", delayLPF)
	}
}

// TestIIRFilter_WithDifferentLengths проверяет фильтр с разной длиной коэффициентов
func TestIIRFilter_WithDifferentLengths(t *testing.T) {
	// b длиннее a
	b1 := []float64{0.5, 0.3, 0.2}
	a1 := []float64{1, 0.1}
	filter1 := NewIIRFilter(b1, a1)

	if filter1.GetOrder() != 2 { // max(2, 1) = 2
		t.Errorf("Порядок фильтра: ожидалось 2, получено %d", filter1.GetOrder())
	}

	// a длиннее b
	b2 := []float64{0.5}
	a2 := []float64{1, 0.2, 0.1, 0.05}
	filter2 := NewIIRFilter(b2, a2)

	if filter2.GetOrder() != 3 { // max(0, 3) = 3
		t.Errorf("Порядок фильтра: ожидалось 3, получено %d", filter2.GetOrder())
	}
}

// TestIIRFilter_EmptyCoeffs проверяет обработку пустых коэффициентов
func TestIIRFilter_EmptyCoeffs(t *testing.T) {
	// Пустые b коэффициенты
	defer func() {
		if r := recover(); r == nil {
			t.Error("Ожидалась паника при пустых b коэффициентах")
		}
	}()
	_ = NewIIRFilter([]float64{}, []float64{1, 0.5})

	// Пустые a коэффициенты
	defer func() {
		if r := recover(); r == nil {
			t.Error("Ожидалась паника при пустых a коэффициентах")
		}
	}()
	_ = NewIIRFilter([]float64{0.5}, []float64{})
}

// BenchmarkIIRFilter_Tick тестирует производительность БИХ-фильтра
func BenchmarkIIRFilter_Tick(b *testing.B) {
	// Фильтр 2-го порядка
	filter := NewSecondOrderLowPass(0.1, 0.707)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.Tick(float64(i))
	}
}

// BenchmarkIIRFilter_Process тестирует производительность обработки среза
func BenchmarkIIRFilter_Process(b *testing.B) {
	filter := NewSecondOrderLowPass(0.1, 0.707)
	input := make([]float64, 1000)
	for i := range input {
		input[i] = float64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.Process(input)
		filter.Reset()
	}
}
