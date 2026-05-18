package proxy

import "fmt"

func installPage(host string) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="ru">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Rebellion сертификат</title>
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
      text-align: center;
    }
    h1 { margin: 0 0 28px; font-size: clamp(28px, 5vw, 48px); }
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
    }
    a:hover { transform: translateY(-2px); background: #bfdbfe; }
    @media (max-width: 560px) { .buttons { grid-template-columns: 1fr; } main { padding: 24px; } }
  </style>
</head>
<body>
  <main>
    <h1>Установите Rebellion сертификат</h1>
    <div class="buttons">
      <a href="http://%s/android.cer">Android сертификат</a>
      <a href="http://%s/ios.pem">iOS сертификат</a>
    </div>
  </main>
</body>
</html>`, host, host)
}
