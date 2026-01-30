# DSP Go Library

Библиотека компонентов для цифровой обработки сигналов. 
Собственная реализация.

## Установка
```bash
go get github.com/Alexxtn105/dsp
```

## Тестирование
### Запуск всех тестов
```bash
go test ./...
```
### Запуск всех с подробным выводом
```bash
go test -v ./...
```

### Запуск только КИХ-тестов
```bash
go test ./Filters/... -run "FIR"
```

### Запуск только БИХ-тестов
```bash
go test ./Filters/... -run "IIR"
```

### Запуск только тестов фильтра Герцеля
```bash
go test ./Filters/... -run "Goertzel"
```

### Запуск только генераторов
```bash
go test ./generators/...
```

### Запуск только генераторов
```bash
go test ./generators/... -run "TestInfo"
```

### Запуск только детекторов
```bash
go test ./detectors/...
```

### Запуск с покрытием кода
```bash
go test -cover ./...
```

### Запуск бенчмарков
```bash
go test -bench=./...
```

### Генерация godoc
```bash
godoc -http=:6060
```

### Просмотр тестового покрытия
```bash
go test -coverprofile=coverage.out ./Filters/
```

```bash
go tool cover -html=coverage.out
```