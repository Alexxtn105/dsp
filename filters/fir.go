package filters

// FIRFilter представляет собой структуру КИХ-фильтра
type FIRFilter struct {
	coeffs []float64 // Коэффициенты фильтра
	buffer []float64 // Кольцевой буфер задержанных отсчетов сигнала
	pos    int       // Текущая позиция в буфере
}

// NewFIRFilter создает новый экземпляр фильтра, принимая массив коэффициентов
func NewFIRFilter(coeffs []float64) *FIRFilter {
	if len(coeffs) == 0 {
		panic("FIRFilter: coefficients cannot be empty")
	}

	n := len(coeffs)
	return &FIRFilter{
		coeffs: coeffs,
		buffer: make([]float64, n),
		pos:    n - 1, // pos указывает на позицию для нового элемента
	}
}

// Tick применяет фильтр к одному новому отсчету
func (f *FIRFilter) Tick(input float64) float64 {
	// Перемещаем позицию и записываем новый отсчет
	f.pos = (f.pos + 1) % len(f.buffer)
	f.buffer[f.pos] = input

	// Вычисляем свертку
	var output float64
	coeffIdx := 0
	bufIdx := f.pos

	for coeffIdx < len(f.coeffs) {
		output += f.coeffs[coeffIdx] * f.buffer[bufIdx]
		coeffIdx++

		// Двигаемся назад по буферу
		bufIdx--
		if bufIdx < 0 {
			bufIdx = len(f.buffer) - 1
		}
	}

	return output
}

// Reset сбрасывает состояние фильтра (очищает буфер)
func (f *FIRFilter) Reset() {
	for i := range f.buffer {
		f.buffer[i] = 0
	}
	f.pos = len(f.buffer) - 1
}

// GetCoefficients возвращает копию коэффициентов фильтра
func (f *FIRFilter) GetCoefficients() []float64 {
	coeffs := make([]float64, len(f.coeffs))
	copy(coeffs, f.coeffs)
	return coeffs
}

// GetBufferSize возвращает размер буфера фильтра
func (f *FIRFilter) GetBufferSize() int {
	return len(f.buffer)
}
