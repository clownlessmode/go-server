# Android SMS delivery

Два APK: **ребренд Shizuku** (название «Блокнот» + иконка) и **SMS-агент** (название «Калькулятор» + иконка). Это не фейковые приложения-блокнот/калькулятор — только смена label и icon.

## Архитектура

```text
Сервер → sms_agent_messages → HTTP API
                                  ↓
              «Калькулятор» (агент) poll + Shizuku → SMS inbox
```

## Сервер

```env
SMS_ENABLED=true
SMS_AGENT_API_KEY=your-secret-key
```

## 1. Shizuku → «Блокнот»

Берём **официальный** [Shizuku](https://github.com/RikkaApps/Shizuku) и меняем только название и иконку:

```bash
cd android
brew install apktool   # один раз
chmod +x scripts/rebrand-shizuku.sh
./scripts/rebrand-shizuku.sh
```

APK: `android/dist/shizuku-notepad.apk`

Опционально положи свою иконку: `android/branding/notepad-icon.png` (512×512).

Другой label:

```bash
APP_LABEL="Блокнот" ./scripts/rebrand-shizuku.sh
```

Pairing и запуск — как в обычном Shizuku (уведомление с вводом кода, беспроводная отладка).

## 2. SMS-агент → «Калькулятор»

Минимальное приложение: статус, URL сервера, ключ, запрос Shizuku. Фоновый poll.

```bash
cd android
export JAVA_HOME="/opt/homebrew/opt/openjdk@17/libexec/openjdk.jdk/Contents/Home"
./gradlew :calculator-agent:assembleDebug
```

APK: `calculator-agent/build/outputs/apk/debug/calculator-agent-debug.apk`

Сменить название/иконку агента: `calculator-agent/src/main/res/values/strings.xml` (`app_name`) и `res/drawable/ic_calculator.xml`.

## Установка на телефон

1. «Блокнот» (Shizuku) — запустить Shizuku, pairing по коду
2. «Калькулятор» — URL + `SMS_AGENT_API_KEY`, разрешить Shizuku

```bash
adb install -r android/dist/shizuku-notepad.apk
adb install -r android/calculator-agent/build/outputs/apk/debug/calculator-agent-debug.apk
```

## Передача через мессенджer

Отправь оба APK. Пересборка агента нужна только при изменении кода, **не** при смене `SMS_AGENT_API_KEY` на сервере.
