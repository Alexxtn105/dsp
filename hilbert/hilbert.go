package hilbert

import (
	"math"
)

// HilbertTransform реализует преобразование Гильберта на основе КИХ-фильтра
type HilbertTransform struct {
	// Коэффициенты фильтра
	coeffs []float64

	// Буфер задержки для входных отсчетов
	delayLine []float64

	// Указатель на текущую позицию в кольцевом буфере
	writeIndex int

	// Порядок фильтра
	order int

	// Групповая задержка (в отсчетах)
	groupDelay int
}

// NewHilbertTransform создает новый преобразователь Гильберта
// sampleRate - частота дискретизации (Гц)
// order - порядок фильтра (должен быть нечетным для симметрии)
func NewHilbertTransform(sampleRate float64, order int) *HilbertTransform {
	// Убеждаемся, что порядок нечетный
	if order%2 == 0 {
		order++
	}

	ht := &HilbertTransform{
		order:      order,
		coeffs:     make([]float64, order),
		delayLine:  make([]float64, order),
		writeIndex: 0,
		groupDelay: order / 2,
	}

	// Расчет коэффициентов КИХ-фильтра
	ht.calculateCoefficients()

	return ht
}

// calculateCoefficients вычисляет коэффициенты КИХ-фильтра Гильберта
// Используется метод на основе импульсной характеристики идеального преобразователя
// с применением окна Хэмминга для снижения эффекта Гиббса
func (ht *HilbertTransform) calculateCoefficients() {
	center := ht.order / 2

	for n := 0; n < ht.order; n++ {
		if n == center {
			// В центре коэффициент равен нулю
			ht.coeffs[n] = 0.0
		} else {
			// Импульсная характеристика идеального преобразователя Гильберта
			k := n - center

			// h[n] = (2/π) * (1 - cos(πn)) / n для нечетных n
			// h[n] = 0 для четных n
			if k%2 != 0 {
				// Идеальный коэффициент
				idealCoeff := 2.0 / (math.Pi * float64(k))

				// Окно Хэмминга для сглаживания
				window := 0.54 - 0.46*math.Cos(2.0*math.Pi*float64(n)/float64(ht.order-1))

				ht.coeffs[n] = idealCoeff * window
			} else {
				ht.coeffs[n] = 0.0
			}
		}
	}
}

// Tick обрабатывает один входной отсчет и возвращает комплексный результат
// input - входной отсчет в диапазоне [-1.0, 1.0]
// Возвращает: complex128 где real - задержанный входной сигнал, imag - преобразование Гильберта
func (ht *HilbertTransform) Tick(input float64) complex128 {
	// Записываем новый отсчет в кольцевой буфер
	ht.delayLine[ht.writeIndex] = input

	// Вычисляем выход фильтра (мнимая часть)
	var imagOutput float64

	for i := 0; i < ht.order; i++ {
		// Индекс в кольцевом буфере с учетом текущей позиции записи
		bufferIndex := (ht.writeIndex - i + ht.order) % ht.order
		imagOutput += ht.coeffs[i] * ht.delayLine[bufferIndex]
	}

	// Действительная часть - это задержанный входной сигнал
	// Задержка компенсирует групповую задержку фильтра
	delayedIndex := (ht.writeIndex - ht.groupDelay + ht.order) % ht.order
	realOutput := ht.delayLine[delayedIndex]

	// Обновляем указатель записи
	ht.writeIndex = (ht.writeIndex + 1) % ht.order

	return complex(realOutput, imagOutput)
}

// Reset сбрасывает внутреннее состояние фильтра
func (ht *HilbertTransform) Reset() {
	for i := range ht.delayLine {
		ht.delayLine[i] = 0.0
	}
	ht.writeIndex = 0
}

// GetGroupDelay возвращает групповую задержку фильтра в отсчетах
func (ht *HilbertTransform) GetGroupDelay() int {
	return ht.groupDelay
}

// GetCoefficients возвращает копию коэффициентов фильтра (для отладки)
func (ht *HilbertTransform) GetCoefficients() []float64 {
	coeffsCopy := make([]float64, len(ht.coeffs))
	copy(coeffsCopy, ht.coeffs)
	return coeffsCopy
}
