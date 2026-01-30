package filters

import (
	"fmt"
	"math"
)

// GoertzelFilter представляет собой структуру фильтра Герцеля для выявления одной частоты
type GoertzelFilter struct {
	k      int     // Частота отсчёта, соответствующая искомой частоте
	w      float64 // Угловой коэффициент
	cosW   float64 // Косинус углового коэффициента
	sinW   float64 // Синус углового коэффициента
	q1     float64 // Состояние q[n-1]
	q2     float64 // Состояние q[n-2]
	n      int     // Текущий отсчёт
	totalN int     // Полное количество выборок для анализа
	coeff  float64 // Коэффициент для рекуррентной формулы: 2*cos(w)
}

// NewGoertzelFilter создает новый экземпляр фильтра Герцеля
func NewGoertzelFilter(freq float64, samplingRate float64, totalN int) (*GoertzelFilter, error) {
	// Проверка граничных условий
	if freq <= 0 {
		return nil, &InvalidParameterError{Param: "freq", Value: freq, Reason: "frequency must be positive"}
	}
	if samplingRate <= 0 {
		return nil, &InvalidParameterError{Param: "samplingRate", Value: samplingRate, Reason: "sampling rate must be positive"}
	}
	if totalN <= 0 {
		return nil, &InvalidParameterError{Param: "totalN", Value: float64(totalN), Reason: "total samples must be positive"}
	}
	if freq >= samplingRate/2 {
		return nil, &InvalidParameterError{
			Param:  "freq",
			Value:  freq,
			Reason: "frequency must be less than Nyquist frequency (samplingRate/2)",
		}
	}

	// Расчет параметров
	k := int(0.5 + float64(totalN)*freq/samplingRate)
	if k >= totalN {
		k = totalN - 1 // Ограничение по теореме Котельникова
	}

	w := 2 * math.Pi * float64(k) / float64(totalN)
	cosW := math.Cos(w)
	sinW := math.Sin(w)

	return &GoertzelFilter{
		k:      k,
		w:      w,
		cosW:   cosW,
		sinW:   sinW,
		coeff:  2 * cosW,
		q1:     0,
		q2:     0,
		n:      0,
		totalN: totalN,
	}, nil
}

// Process обрабатывает одно значение сигнала и накапливает состояние фильтра
func (gf *GoertzelFilter) Process(input float64) error {
	if gf == nil {
		return &InvalidStateError{Reason: "filter is not initialized"}
	}

	if gf.n >= gf.totalN {
		return &InvalidStateError{Reason: "all samples have already been processed"}
	}

	// Основное рекуррентное соотношение фильтра Герцеля:
	// q[n] = x[n] + coeff * q[n-1] - q[n-2]
	q0 := input + gf.coeff*gf.q1 - gf.q2

	// Сдвигаем состояния
	gf.q2 = gf.q1
	gf.q1 = q0
	gf.n++

	return nil
}

// Reset сбрасывает состояние фильтра для нового расчета
func (gf *GoertzelFilter) Reset() error {
	if gf == nil {
		return &InvalidStateError{Reason: "filter is not initialized"}
	}

	gf.q1 = 0
	gf.q2 = 0
	gf.n = 0
	return nil
}

// GetMagnitude возвращает амплитуду найденной частоты
func (gf *GoertzelFilter) GetMagnitude() (float64, error) {
	if gf == nil {
		return 0, &InvalidStateError{Reason: "filter is not initialized"}
	}

	if gf.n == 0 {
		return 0, &InvalidStateError{Reason: "no samples have been processed yet"}
	}

	// Правильная формула для вычисления амплитуды в фильтре Герцеля:
	// magnitude = sqrt(q1^2 + q2^2 - coeff * q1 * q2) * (2/N)
	// Где q1 = q[N-1], q2 = q[N-2], coeff = 2*cos(w)

	magnitudeSquared := gf.q1*gf.q1 + gf.q2*gf.q2 - gf.coeff*gf.q1*gf.q2

	if magnitudeSquared < 0 {
		// Из-за погрешностей вычислений с плавающей точкой
		return 0, nil
	}

	// Важно: здесь мы используем 2/float64(gf.totalN) для нормировки
	magnitude := 2 * math.Sqrt(magnitudeSquared) / float64(gf.totalN)

	return magnitude, nil
}

// GetMagnitudeOptimized возвращает амплитуду с оптимизированной формулой
func (gf *GoertzelFilter) GetMagnitudeOptimized() (float64, error) {
	if gf == nil {
		return 0, &InvalidStateError{Reason: "filter is not initialized"}
	}

	if gf.n == 0 {
		return 0, &InvalidStateError{Reason: "no samples have been processed yet"}
	}

	// Альтернативная оптимизированная формула
	// real = q1 - q2*cos(w)
	// imag = q2*sin(w)
	// magnitude = sqrt(real^2 + imag^2) * (2/N)
	realPart := gf.q1 - gf.q2*gf.cosW
	imagPart := gf.q2 * gf.sinW
	magnitudeSquared := realPart*realPart + imagPart*imagPart

	if magnitudeSquared < 0 {
		return 0, nil
	}

	// Нормировка такая же, как в GetMagnitude
	magnitude := 2 * math.Sqrt(magnitudeSquared) / float64(gf.totalN)

	return magnitude, nil
}

// GetPower возвращает мощность сигнала на целевой частоте
func (gf *GoertzelFilter) GetPower() (float64, error) {
	magnitude, err := gf.GetMagnitude()
	if err != nil {
		return 0, err
	}
	return magnitude * magnitude / 2, nil
}

// IsComplete возвращает true, если обработаны все выборки
func (gf *GoertzelFilter) IsComplete() bool {
	if gf == nil {
		return false
	}
	return gf.n >= gf.totalN
}

// GetProcessedCount возвращает количество обработанных отсчетов
func (gf *GoertzelFilter) GetProcessedCount() int {
	if gf == nil {
		return 0
	}
	return gf.n
}

// GetTargetFrequency возвращает целевую частоту
func (gf *GoertzelFilter) GetTargetFrequency(samplingRate float64) float64 {
	if gf == nil || gf.totalN == 0 {
		return 0
	}
	return float64(gf.k) * samplingRate / float64(gf.totalN)
}

// GetCoefficient возвращает коэффициент k
func (gf *GoertzelFilter) GetCoefficient() int {
	if gf == nil {
		return 0
	}
	return gf.k
}

// InvalidParameterError представляет ошибку неверного параметра
type InvalidParameterError struct {
	Param  string
	Value  float64
	Reason string
}

func (e *InvalidParameterError) Error() string {
	return fmt.Sprintf("invalid parameter %s: %f - %s", e.Param, e.Value, e.Reason)
}

// InvalidStateError представляет ошибку неверного состояния
type InvalidStateError struct {
	Reason string
}

func (e *InvalidStateError) Error() string {
	return fmt.Sprintf("invalid filter state: %s", e.Reason)
}
