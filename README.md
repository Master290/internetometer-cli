# internetometer-cli > Яндекс Интернетометр в терминале
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) 

Неофициальный CLI для Яндекс Интернетометр (yandex.ru/internet) 
![скриншот](https://i.imgur.com/gz0yhdw.png)

  - Определение IPv4 и IPv6.
  - Определение региона, провайдера и номера автономной системы (ISP/ASN).
  - Точное измерение задержки (Ping).
- **Различные форматы вывода**:
  - Читаемый текстовый формат
  - JSON
  - Экспорт метрик Prometheus
  - JSONL

## Установка
> Готовые бинарники лежат в [релизах](https://github.com/Master290/internetometer-cli/releases)

Убедитесь, что у вас установлен [Go](https://go.dev/doc/install).

```bash
git clone https://github.com/Master290/internetometer-cli.git
cd internetometer-cli
go mod tidy
# CLI
go build -o internetometer ./cmd/cli/main.go
# экспортер
go build -o prom-exporter ./cmd/prom/exporter.go
```

## Установка (Arch Linux / AUR)

Если вы используете Arch Linux, вы можете установить пакеты напрямую из AUR:

**CLI версия:**
```bash
yay -S internetometer-cli
```

**Prometheus Exporter:**
```bash
yay -S internetometer-exporter
```

## Быстрый старт

### CLI (Консольная версия)
```bash
./internetometer
```

### Экспортер Prometheus (Фоновый режим)[*](https://github.com/Master290/internetometer-cli/pull/5)
Запуск HTTP-сервера с метриками (по умолчанию на :9112):
```bash
./prom-exporter --delay 1h
```

### Основные флаги

- `--speed`: Просто текстовый режим, без красивого TUI.
- `--all`: Подробный вывод: IPv4/6, регион, ISP, вход./исход. скорости, задержка, ОС и время.
- `--json`: Вывод в формате JSON.
- `--lang ru`: Использовать русский язык, так же есть вариант `--lang en` для английского языка. (пока что только меняет название региона)
- `--save log.jsonl`: Сохранить результат в лог-файл.
- `--prometheus`: Вывод в формате метрик Prometheus.
- `--concurrency 4`: Количество параллельных потоков.
