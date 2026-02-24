# internetometer-cli > Яндекс Интернетометр в терминале
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Неофициальный CLI для Яндекс Интернетометр (yandex.ru/internet) 

  - Определение IPv4 и IPv6.
  - Определение региона.
  - Определение провайдера и номера автономной системы (ISP/ASN).
  - Точное измерение задержки (Ping).
- **Различные форматы вывода**:
  - Читаемый текстовый формат
  - JSON
  - Экспорт метрик Prometheus
  - JSONL

## Установка

Убедитесь, что у вас установлен [Go](https://go.dev/doc/install).

```bash
git clone https://github.com/Master290/internetometer-cli.git
cd internetometer-cli
go build -o internetometer main.go
```

## Быстрый старт

Просто запустите программу без флагов:
```bash
./internetometer
```

### Основные флаги

- `--speed`: Просто текстовый режим, без красивого TUI.
- `--all`: Подробный вывод: IPv4/6, регион, ISP, вход./исход. скорости, задержка, ОС и время.
- `--json`: Вывод в формате JSON.
- `--lang ru`: Использовать русский язык, так же есть вариант `--lang en` для английского языка. (пока что только меняет название региона)
- `--save log.jsonl`: Сохранить результат в лог-файл.
- `--prometheus`: Вывод в формате метрик Prometheus.
- `--concurrency 4`: Количество параллельных потоков.
