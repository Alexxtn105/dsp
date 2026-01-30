
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
go test ./filters/... -run "FIR"
```

### Запуск только БИХ-тестов
```bash
go test ./filters/... -run "IIR"
```

### Запуск только тестов фильтра Герцеля
```bash
go test ./filters/... -run "Goertzel"
```

### Запуск только генераторов
```bash
go test ./generators/...
```

### Запуск только окон
```bash
go test ./windows/...
```
### Запуск только окон (каждое отдельно)
```bash
go test ./windows/... -run TestBlackmanHarrisWindow
```
```bash
go test ./windows/... -run TestHammingWindow
```
```bash
go test ./windows/... -run TestHannWindow
```
```bash
go test ./windows/... -run TestKaiserWindow
```
```bash
go test ./windows/... -run TestNatollWindow
```
```bash
go test ./windows/... -run TestTukeyWindow
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
go test -coverprofile=coverage.out ./filters/
```

```bash
go tool cover -html=coverage.out
```