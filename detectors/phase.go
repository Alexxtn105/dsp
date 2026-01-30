package detectors

import (
	"math"
	"math/cmplx"
)

// CoherentPhaseDetector представляет собой структуру фазового детектора
type CoherentPhaseDetector struct {
	referenceSignal complex128 // Опорный сигнал (нормированный)
	phaseOffset     float64    // Компенсационное смещение фазы
	alpha           float64    // Коэффициент фильтрации (0 < alpha <= 1)
	filteredError   float64    // Отфильтрованная ошибка фазы
}

// NewCoherentPhaseDetector создает новый экземпляр фазового детектора
func NewCoherentPhaseDetector(referenceSignal complex128, alpha float64) *CoherentPhaseDetector {
	// Нормируем опорный сигнал
	refMagnitude := cmplx.Abs(referenceSignal)
	refNorm := referenceSignal / complex(refMagnitude, 0)

	if alpha <= 0 || alpha > 1 {
		alpha = 0.1 // значение по умолчанию
	}

	return &CoherentPhaseDetector{
		referenceSignal: refNorm,
		phaseOffset:     0,
		alpha:           alpha,
		filteredError:   0,
	}
}

// Detect измеряет и фильтрует ошибку фазы
func (cpd *CoherentPhaseDetector) Detect(inputSignal complex128) float64 {
	// Нормируем входной сигнал
	inputMagnitude := cmplx.Abs(inputSignal)
	inputNorm := inputSignal / complex(inputMagnitude, 0)

	// Вычисляем разность фаз
	phaseDiff := cmplx.Phase(inputNorm) - cmplx.Phase(cpd.referenceSignal)

	// Нормализуем разность фаз в диапазон [-π, π]
	phaseDiff = normalizePhase(phaseDiff)

	// Применяем фильтр низких частот (петлевой фильтр)
	cpd.filteredError = cpd.alpha*phaseDiff + (1-cpd.alpha)*cpd.filteredError

	// Корректируем с учетом текущего смещения
	correctedPhase := cpd.filteredError - cpd.phaseOffset

	// Нормализуем результат
	return normalizePhase(correctedPhase)
}

// UpdateOffset обновляет смещение фазы на основе текущей ошибки
func (cpd *CoherentPhaseDetector) UpdateOffset() {
	// Используем отфильтрованную ошибку для коррекции
	cpd.phaseOffset += cpd.filteredError
	cpd.filteredError = 0 // Сбрасываем после коррекции
}

// SetPhaseOffset устанавливает конкретное смещение фазы
func (cpd *CoherentPhaseDetector) SetPhaseOffset(offset float64) {
	cpd.phaseOffset = normalizePhase(offset)
}

//// normalizePhase нормализует фазу в диапазон [-π, π]
//func normalizePhase(phase float64) float64 {
//	// Приводим фазу к диапазону [-π, π]
//	phase = math.Mod(phase+math.Pi, 2*math.Pi)
//	if phase < 0 {
//		phase += 2 * math.Pi
//	}
//	return phase - math.Pi
//}

// normalizePhase нормализует фазу в диапазон [-π, π]
func normalizePhase(phase float64) float64 {
	// Приводим фазу к диапазону [-π, π]
	phase = math.Mod(phase+math.Pi, 2*math.Pi)
	if phase < 0 {
		phase += 2 * math.Pi
	}
	// Преобразуем из [0, 2π] в [-π, π]
	phase -= math.Pi

	// Особый случай: если фаза равна -π, возвращаем π для согласованности
	if phase == -math.Pi {
		return math.Pi
	}
	return phase
}

// GetFilteredError возвращает текущую отфильтрованную ошибку
func (cpd *CoherentPhaseDetector) GetFilteredError() float64 {
	return cpd.filteredError
}

// GetPhaseOffset возвращает текущее смещение фазы
func (cpd *CoherentPhaseDetector) GetPhaseOffset() float64 {
	return cpd.phaseOffset
}

// UpdateReferenceSignal обновляет опорный сигнал
func (cpd *CoherentPhaseDetector) UpdateReferenceSignal(newRef complex128) {
	magnitude := cmplx.Abs(newRef)
	cpd.referenceSignal = newRef / complex(magnitude, 0)
}
