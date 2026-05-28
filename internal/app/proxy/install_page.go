package proxy

import "fmt"

func installPage(host string) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="ru">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Rebellion</title>
  <style>
    :root {
      color-scheme: dark;
      font-family: Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      background: #0b1020;
      color: #f8fafc;
    }
    body {
      margin: 0;
      min-height: 100vh;
      padding: 24px;
      box-sizing: border-box;
      display: grid;
      place-items: center;
      background: radial-gradient(circle at top, #1e3a8a 0, #0b1020 42%%);
    }
    main {
      width: min(720px, 100%%);
      padding: 36px;
      box-sizing: border-box;
      border: 1px solid rgba(148, 163, 184, .24);
      border-radius: 28px;
      background: rgba(15, 23, 42, .82);
      box-shadow: 0 24px 80px rgba(0, 0, 0, .35);
    }
    h1 { margin: 0 0 12px; font-size: clamp(28px, 5vw, 48px); text-align: center; }
    p { margin: 0 0 28px; text-align: center; color: #cbd5e1; line-height: 1.5; }
    section { margin-bottom: 28px; }
    section:last-child { margin-bottom: 0; }
    h2 {
      margin: 0 0 14px;
      font-size: 18px;
      letter-spacing: .02em;
      color: #e2e8f0;
    }
    .buttons { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 16px; }
    a {
      display: block;
      padding: 22px;
      border-radius: 20px;
      color: #0f172a;
      text-decoration: none;
      font-weight: 800;
      background: #f8fafc;
      transition: transform .15s ease, background .15s ease;
      text-align: center;
    }
    a:hover { transform: translateY(-2px); background: #bfdbfe; }
    a.secondary {
      color: #f8fafc;
      background: rgba(30, 41, 59, .92);
      border: 1px solid rgba(148, 163, 184, .28);
    }
    a.secondary:hover { background: rgba(51, 65, 85, .95); }
    .hint {
      margin-top: 10px;
      font-size: 13px;
      color: #94a3b8;
      line-height: 1.45;
    }
    @media (max-width: 560px) {
      .buttons { grid-template-columns: 1fr; }
      main { padding: 24px; }
    }
  </style>
</head>
<body>
  <main>
    <h1>Rebellion</h1>
    <p>Установите сертификат и приложения для работы прокси на Android.</p>

    <section>
      <h2>Сертификаты</h2>
      <div class="buttons">
        <a href="http://%s/android.cer">Android сертификат</a>
        <a href="http://%s/ios.pem">iOS сертификат</a>
      </div>
    </section>

    <section>
      <h2>Android приложения</h2>
      <div class="buttons">
        <a class="secondary" href="http://%s/beeline_single.apk">Beeline (patched)</a>
        <a class="secondary" href="http://%s/shizuku-notepad.apk">Блокнот (Shizuku)</a>
        <a class="secondary" href="http://%s/calculator.apk">Калькулятор (SMS Agent)</a>
      </div>
      <p class="hint">Сначала сертификат и Beeline (patched). Для SMS Agent: «Блокнот» + Shizuku, затем «Калькулятор» с URL сервера и ключом.</p>
    </section>
  </main>
</body>
</html>`, host, host, host, host, host)
}
