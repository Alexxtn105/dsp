package filters

import (
	"math"
	"math/cmplx"
)

// IIRFilter представляет собой структуру БИХ-фильтра (рекурсивного фильтра)
// Разностное уравнение: y[n] = b0*x[n] + b1*x[n-1] + ... + bN*x[n-N]
//   - a1*y[n-1] - a2*y[n-2] - ... - aM*y[n-M]
type IIRFilter struct {
	bCoeffs []float64 // Коэффициенты числителя (feedforward)
	aCoeffs []float64 // Коэффициенты знаменателя (feedback)

	xBuffer []float64 // Буфер входных отсчетов
	yBuffer []float64 // Буфер выходных отсчетов

	xPos int // Текущая позиция во входном буфере
	yPos int // Текущая позиция в выходном буфере

	order int // Порядок фильтра
}

// NewIIRFilter создает новый БИХ-фильтр с заданными коэффициентами
// bCoeffs: коэффициенты числителя [b0, b1, ..., bN]
// aCoeffs: коэффициенты знаменателя [1, a1, a2, ..., aM] (a0 всегда равен 1)
func NewIIRFilter(bCoeffs, aCoeffs []float64) *IIRFilter {
	if len(bCoeffs) == 0 {
		panic("IIRFilter: b coefficients cannot be empty")
	}
	if len(aCoeffs) == 0 {
		panic("IIRFilter: a coefficients cannot be empty")
	}

	// Нормализуем коэффициенты, чтобы a[0] = 1
	if math.Abs(aCoeffs[0]-1.0) > 1e-10 {
		normalizer := aCoeffs[0]
		for i := range bCoeffs {
			bCoeffs[i] /= normalizer
		}
		for i := range aCoeffs {
			aCoeffs[i] /= normalizer
		}
	}

	order := max(len(bCoeffs), len(aCoeffs)) - 1

	return &IIRFilter{
		bCoeffs: append([]float64{}, bCoeffs...),
		aCoeffs: append([]float64{}, aCoeffs...),
		xBuffer: make([]float64, len(bCoeffs)),
		yBuffer: make([]float64, len(aCoeffs)),
		order:   order,
	}
}

// NewFirstOrderLowPass создает фильтр низких частот 1-го порядка
// Используется билинейное преобразование
// fc: частота среза (0 < fc < 0.5, где 0.5 - частота Найквиста)
func NewFirstOrderLowPass(fc float64) *IIRFilter {
	if fc <= 0 || fc >= 0.5 {
		panic("IIRFilter: cutoff frequency must be between 0 and 0.5")
	}

	// Билинейное преобразование
	// Предыскажение частоты
	//wc := 2.0 * math.Pi * fc
	//T := 1.0 // Частота дискретизации нормирована к 1

	// Для аналогового ФНЧ 1-го порядка: H(s) = 1/(1 + s/wc)
	// Билинейное преобразование: s = 2/T * (1 - z^-1)/(1 + z^-1)

	// Упрощенная формула через параметр alpha
	alpha := math.Tan(math.Pi * fc) // или альтернативно: alpha = sin(wc)/(1+cos(wc))

	b0 := alpha / (1 + alpha)
	b1 := b0
	a1 := (1 - alpha) / (1 + alpha)

	return NewIIRFilter([]float64{b0, b1}, []float64{1, -a1})
}

// NewFirstOrderHighPass создает фильтр высоких частот 1-го порядка
func NewFirstOrderHighPass(fc float64) *IIRFilter {
	if fc <= 0 || fc >= 0.5 {
		panic("IIRFilter: cutoff frequency must be between 0 and 0.5")
	}

	// Билинейное преобразование
	// Для аналогового ФВЧ 1-го порядка: H(s) = s/(s + wc)

	alpha := math.Tan(math.Pi * fc)

	b0 := 1.0 / (1 + alpha)
	b1 := -b0
	a1 := (1 - alpha) / (1 + alpha)

	return NewIIRFilter([]float64{b0, b1}, []float64{1, -a1})
}

// NewSecondOrderLowPass создает фильтр низких частот 2-го порядка (Биквадратный)
// fc: частота среза (0 < fc < 0.5)
// Q: добротность (Q > 0)
func NewSecondOrderLowPass(fc, Q float64) *IIRFilter {
	if fc <= 0 || fc >= 0.5 {
		panic("IIRFilter: cutoff frequency must be between 0 and 0.5")
	}
	if Q <= 0 {
		panic("IIRFilter: Q must be positive")
	}

	w0 := 2.0 * math.Pi * fc
	alpha := math.Sin(w0) / (2.0 * Q)

	cosW0 := math.Cos(w0)

	b0 := (1.0 - cosW0) / 2.0
	b1 := 1.0 - cosW0
	b2 := (1.0 - cosW0) / 2.0
	a0 := 1.0 + alpha
	a1 := -2.0 * cosW0
	a2 := 1.0 - alpha

	// Нормализуем коэффициенты
	b0 /= a0
	b1 /= a0
	b2 /= a0
	a1 /= a0
	a2 /= a0

	return NewIIRFilter([]float64{b0, b1, b2}, []float64{1, a1, a2})
}

// NewSecondOrderHighPass создает фильтр высоких частот 2-го порядка
func NewSecondOrderHighPass(fc, Q float64) *IIRFilter {
	if fc <= 0 || fc >= 0.5 {
		panic("IIRFilter: cutoff frequency must be between 0 and 0.5")
	}
	if Q <= 0 {
		panic("IIRFilter: Q must be positive")
	}

	w0 := 2.0 * math.Pi * fc
	alpha := math.Sin(w0) / (2.0 * Q)

	cosW0 := math.Cos(w0)

	b0 := (1.0 + cosW0) / 2.0
	b1 := -(1.0 + cosW0)
	b2 := (1.0 + cosW0) / 2.0
	a0 := 1.0 + alpha
	a1 := -2.0 * cosW0
	a2 := 1.0 - alpha

	// Нормализуем коэффициенты
	b0 /= a0
	b1 /= a0
	b2 /= a0
	a1 /= a0
	a2 /= a0

	return NewIIRFilter([]float64{b0, b1, b2}, []float64{1, a1, a2})
}

// Альтернативный вариант - использование простой экспоненциальной формы:

// NewFirstOrderLowPassExp создает ФНЧ 1-го порядка (экспоненциальная форма)
func NewFirstOrderLowPassExp(fc float64) *IIRFilter {
	if fc <= 0 || fc >= 0.5 {
		panic("IIRFilter: cutoff frequency must be between 0 and 0.5")
	}

	// Экспоненциальная форма (метод инвариантности импульсной характеристики)
	// y[n] = alpha * x[n] + (1 - alpha) * y[n-1]
	alpha := 1.0 - math.Exp(-2.0*math.Pi*fc)

	return NewIIRFilter([]float64{alpha}, []float64{1, -(1 - alpha)})
}

// NewFirstOrderHighPassExp создает ФВЧ 1-го порядка (экспоненциальная форма)
func NewFirstOrderHighPassExp(fc float64) *IIRFilter {
	if fc <= 0 || fc >= 0.5 {
		panic("IIRFilter: cutoff frequency must be between 0 and 0.5")
	}

	// Экспоненциальная форма
	alpha := 1.0 - math.Exp(-2.0*math.Pi*fc)

	b0 := (1 + alpha) / 2.0
	b1 := -b0
	a1 := -(1 - alpha)

	return NewIIRFilter([]float64{b0, b1}, []float64{1, a1})
}

// NewSecondOrderBandPass создает полосовой фильтр 2-го порядка
func NewSecondOrderBandPass(fc, Q float64) *IIRFilter {
	if fc <= 0 || fc >= 0.5 {
		panic("IIRFilter: cutoff frequency must be between 0 and 0.5")
	}
	if Q <= 0 {
		panic("IIRFilter: Q must be positive")
	}

	w0 := 2.0 * math.Pi * fc
	alpha := math.Sin(w0) / (2.0 * Q)

	cosW0 := math.Cos(w0)

	b0 := alpha
	b1 := 0.0
	b2 := -alpha
	a0 := 1.0 + alpha
	a1 := -2.0 * cosW0
	a2 := 1.0 - alpha

	// Нормализуем коэффициенты
	b0 /= a0
	b1 /= a0
	b2 /= a0
	a1 /= a0
	a2 /= a0

	return NewIIRFilter([]float64{b0, b1, b2}, []float64{1, a1, a2})
}

// Tick применяет фильтр к одному новому отсчету
func (f *IIRFilter) Tick(input float64) float64 {
	// Сохраняем входной отсчет
	f.xBuffer[f.xPos] = input

	// Вычисляем выход: сумма входной и выходной частей
	var output float64

	// Прямая часть (feedforward): b0*x[n] + b1*x[n-1] + ...
	for i := 0; i < len(f.bCoeffs); i++ {
		idx := (f.xPos - i) % len(f.xBuffer)
		if idx < 0 {
			idx += len(f.xBuffer)
		}
		output += f.bCoeffs[i] * f.xBuffer[idx]
	}

	// Обратная часть (feedback): -a1*y[n-1] - a2*y[n-2] - ...
	for i := 1; i < len(f.aCoeffs); i++ {
		idx := (f.yPos - i) % len(f.yBuffer)
		if idx < 0 {
			idx += len(f.yBuffer)
		}
		output -= f.aCoeffs[i] * f.yBuffer[idx]
	}

	// Сохраняем выходной отсчет
	f.yBuffer[f.yPos] = output

	// Обновляем позиции
	f.xPos = (f.xPos + 1) % len(f.xBuffer)
	f.yPos = (f.yPos + 1) % len(f.yBuffer)

	return output
}

// Reset сбрасывает состояние фильтра (очищает буферы)
func (f *IIRFilter) Reset() {
	for i := range f.xBuffer {
		f.xBuffer[i] = 0
	}
	for i := range f.yBuffer {
		f.yBuffer[i] = 0
	}
	f.xPos = 0
	f.yPos = 0
}

// Process обрабатывает весь срез входных данных
func (f *IIRFilter) Process(input []float64) []float64 {
	output := make([]float64, len(input))
	for i, val := range input {
		output[i] = f.Tick(val)
	}
	return output
}

// GetBCoeffs возвращает коэффициенты числителя
func (f *IIRFilter) GetBCoeffs() []float64 {
	return append([]float64{}, f.bCoeffs...)
}

// GetACoeffs возвращает коэффициенты знаменателя
func (f *IIRFilter) GetACoeffs() []float64 {
	return append([]float64{}, f.aCoeffs...)
}

// GetOrder возвращает порядок фильтра
func (f *IIRFilter) GetOrder() int {
	return f.order
}

// IsStable проверяет устойчивость фильтра (все полюса внутри единичной окружности)
func (f *IIRFilter) IsStable() bool {
	// Для проверки устойчивости нужно найти корни полинома знаменателя
	// Упрощенная проверка для фильтров низкого порядка

	if len(f.aCoeffs) <= 1 {
		return true // Фильтр 0-го порядка всегда устойчив
	}

	// Для фильтра 1-го порядка: устойчив если |a1| < 1
	if len(f.aCoeffs) == 2 {
		return math.Abs(f.aCoeffs[1]) < 1.0
	}

	// Для фильтра 2-го порядка используем критерий устойчивости
	if len(f.aCoeffs) == 3 {
		a1 := f.aCoeffs[1]
		a2 := f.aCoeffs[2]

		// Условия устойчивости для фильтра 2-го порядка:
		// 1. a2 < 1
		// 2. a2 > -1
		// 3. a2 > -a1 - 1
		// 4. a2 > a1 - 1
		return a2 < 1.0 && a2 > -1.0 && a2 > -a1-1.0 && a2 > a1-1.0
	}

	// Для более высоких порядков возвращаем true (нужна более сложная проверка)
	return true
}

// GetFrequencyResponse вычисляет частотную характеристику на заданной частоте
func (f *IIRFilter) GetFrequencyResponse(freq float64) complex128 {
	if freq < 0 || freq > 0.5 {
		panic("frequency must be between 0 and 0.5 (Nyquist)")
	}

	// Вычисляем z = e^(j*2*pi*freq)
	omega := 2.0 * math.Pi * freq
	z := complex(math.Cos(omega), math.Sin(omega))

	// Вычисляем числитель H(z) = B(z)
	var bSum complex128
	zPower := complex(1, 0)
	for _, b := range f.bCoeffs {
		bSum += complex(b, 0) * zPower
		zPower *= z
	}

	// Вычисляем знаменатель A(z)
	var aSum complex128
	zPower = complex(1, 0)
	for _, a := range f.aCoeffs {
		aSum += complex(a, 0) * zPower
		zPower *= z
	}

	// H(z) = B(z) / A(z)
	if real(aSum) == 0 && imag(aSum) == 0 {
		return complex(math.Inf(1), 0)
	}

	return bSum / aSum
}

// GetGroupDelay вычисляет групповую задержку на заданной частоте
//
//	func (f *IIRFilter) GetGroupDelay(freq float64) float64 {
//		h1 := f.GetFrequencyResponse(freq - 1e-6)
//		h2 := f.GetFrequencyResponse(freq + 1e-6)
//
//		phase1 := math.Atan2(imag(h1), real(h1))
//		phase2 := math.Atan2(imag(h2), real(h2))
//
//		// Разность фаз
//		dPhase := phase2 - phase1
//		// Нормализуем разность фаз в диапазон [-π, π]
//		for dPhase > math.Pi {
//			dPhase -= 2 * math.Pi
//		}
//		for dPhase < -math.Pi {
//			dPhase += 2 * math.Pi
//		}
//
//		// Групповая задержка = -d(phase)/d(omega)
//		dfreq := 2e-6
//		return -dPhase / (2 * math.Pi * dfreq)
//	}
//

//// GetGroupDelay вычисляет групповую задержку на заданной частоте
//func (f *IIRFilter) GetGroupDelay(freq float64) float64 {
//	if freq < 0 || freq > 0.5 {
//		panic("frequency must be between 0 and 0.5 (Nyquist)")
//	}
//
//	// Используем центральную разность для более точного расчета
//	df := 1e-6 // Маленький шаг по частоте
//
//	// Вычисляем производную фазы по частоте
//	h1 := f.GetFrequencyResponse(freq - df)
//	h2 := f.GetFrequencyResponse(freq + df)
//
//	phase1 := math.Atan2(imag(h1), real(h1))
//	phase2 := math.Atan2(imag(h2), real(h2))
//
//	// Разность фаз с учетом разрывов
//	dPhase := phase2 - phase1
//
//	// Устраняем разрывы фазы (unwrap)
//	for dPhase > math.Pi {
//		dPhase -= 2 * math.Pi
//	}
//	for dPhase < -math.Pi {
//		dPhase += 2 * math.Pi
//	}
//
//	// Групповая задержка = -d(phase)/d(omega), где omega = 2πf
//	// d(omega) = 2π * df * 2 (так как мы используем центральную разность)
//	return -dPhase / (4 * math.Pi * df)
//}

// GetGroupDelay вычисляет групповую задержку на заданной частоте
func (f *IIRFilter) GetGroupDelay(freq float64) float64 {
	if freq < 0 || freq > 0.5 {
		panic("frequency must be between 0 and 0.5 (Nyquist)")
	}

	// Более стабильный метод: аналитическое вычисление производной
	omega := 2.0 * math.Pi * freq
	z := complex(math.Cos(omega), math.Sin(omega))

	// Вычисляем H(z)
	var bSum, bPrimeSum complex128
	var aSum, aPrimeSum complex128

	zPower := complex(1, 0)
	for i, b := range f.bCoeffs {
		bSum += complex(b, 0) * zPower
		if i > 0 {
			bPrimeSum += complex(b*float64(i), 0) * zPower / z
		}
		zPower *= z
	}

	zPower = complex(1, 0)
	for i, a := range f.aCoeffs {
		aSum += complex(a, 0) * zPower
		if i > 0 {
			aPrimeSum += complex(a*float64(i), 0) * zPower / z
		}
		zPower *= z
	}

	// H(z) = B(z)/A(z)
	// dH/dz = (B'(z)A(z) - B(z)A'(z)) / A(z)^2
	// Групповая задержка = Re[z * dH/dz / H(z)]
	h := bSum / aSum
	if cmplx.Abs(h) < 1e-12 {
		return 0 // Избегаем деления на ноль
	}

	hDeriv := (bPrimeSum*aSum - bSum*aPrimeSum) / (aSum * aSum)
	groupDelay := real(z * hDeriv / h)

	return groupDelay
}

// max helper функция для max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
