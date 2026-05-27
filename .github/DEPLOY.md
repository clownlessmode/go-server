# Deploy (GitHub Actions)

Доступ к серверу задаётся **в `.github/workflows/deploy.yml`** (блок `env` вверху):

```yaml
DEPLOY_HOST: "YOUR_SERVER_IP"
DEPLOY_USER: "root"
DEPLOY_PASSWORD: "YOUR_ROOT_PASSWORD"
DEPLOY_PORT: "22"
```

При push в `main` (или **Actions → Deploy to server → Run workflow**):

1. Собирается Linux-бинарник `mitm-api`
2. На сервере обновляется код (`git reset --hard origin/main`)
3. Бинарник копируется в `/opt/mitm-go-server/`
4. `systemctl restart mitm-api`

## Первичная настройка сервера (один раз)

```bash
# зависимости
apt update
apt install -y git postgresql

# клон (подставь URL своего private repo)
git clone git@github.com:YOUR_ORG/go-server.git /opt/mitm-go-server
cd /opt/mitm-go-server

# .env (не в git)
cp .env.example .env 2>/dev/null || nano .env

# systemd
cp deploy/mitm-api.service /etc/systemd/system/mitm-api.service
systemctl daemon-reload
systemctl enable mitm-api postgresql@18-main

# postgres + первый деплой через GitHub Actions
systemctl start postgresql@18-main
```

После этого каждый push в `main` деплоит автоматически.

## Ручной деплой без Actions

```bash
cd /opt/mitm-go-server
git pull
CGO_ENABLED=0 go build -o mitm-api ./cmd/api/main.go
systemctl restart mitm-api
```
