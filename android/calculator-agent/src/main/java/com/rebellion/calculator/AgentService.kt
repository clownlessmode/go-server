package com.rebellion.calculator

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.Service
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.os.IBinder
import androidx.core.app.NotificationCompat
import rikka.shizuku.Shizuku
import java.util.concurrent.Executors
import java.util.concurrent.TimeUnit

class AgentService : Service() {
    private val executor = Executors.newSingleThreadScheduledExecutor()

    override fun onCreate() {
        super.onCreate()
        createChannel()
        startForeground(NOTIFICATION_ID, buildNotification(getString(R.string.agent_running)))
        AgentDiagnostics.setAgentState(this, "работает, опрос каждые ${POLL_INTERVAL_SECONDS} сек")
        AgentDiagnostics.log(this, AgentDiagnostics.Level.INFO, "Агент запущен")
        executor.scheduleWithFixedDelay({ pollOnce() }, 0, POLL_INTERVAL_SECONDS, TimeUnit.SECONDS)
    }

    override fun onDestroy() {
        AgentDiagnostics.setAgentState(this, "остановлен")
        executor.shutdownNow()
        super.onDestroy()
    }

    override fun onBind(intent: Intent?): IBinder? = null

    private fun pollOnce() {
        val config = AgentConfig(this)
        val environment = SmsEnvironment.inspect(this)

        if (!config.isConfigured) {
            AgentDiagnostics.setAgentState(this, "ждёт настройки (URL и ключ)")
            updateNotification(getString(R.string.agent_not_configured))
            return
        }
        if (!environment.shizukuReady) {
            AgentDiagnostics.setAgentState(this, "ждёт Shizuku")
            updateNotification(getString(R.string.agent_waiting_shizuku))
            return
        }

        AgentDiagnostics.setAgentState(this, "работает")

        try {
            val api = AgentApi(config.serverUrl, config.apiKey)
            AgentDiagnostics.log(
                this,
                AgentDiagnostics.Level.INFO,
                "Опрос сервера",
                config.serverUrl.trimEnd('/'),
            )

            val messages = api.fetchPending()
            if (messages.isEmpty()) {
                AgentDiagnostics.setPollResult(this, "пусто — новых SMS нет")
                updateNotification(getString(R.string.agent_idle))
                return
            }

            AgentDiagnostics.setPollResult(this, "получено ${messages.size} SMS")
            AgentDiagnostics.log(
                this,
                AgentDiagnostics.Level.INFO,
                "С сервера пришло SMS: ${messages.size}",
            )

            for (message in messages) {
                deliverMessage(config, environment, api, message)
            }
            updateNotification(getString(R.string.agent_delivered, messages.size))
        } catch (error: Exception) {
            val text = error.message ?: "unknown"
            AgentDiagnostics.setPollResult(this, "ошибка опроса: $text")
            AgentDiagnostics.log(
                this,
                AgentDiagnostics.Level.ERROR,
                "Не удалось опросить сервер",
                text,
            )
            updateNotification(getString(R.string.agent_error, text))
        }
    }

    private fun deliverMessage(
        config: AgentConfig,
        environment: SmsEnvironment.Info,
        api: AgentApi,
        message: PendingMessage,
    ) {
        val bodyPreview = message.body.replace('\n', ' ').take(100)
        try {
            val insertResult = SmsInjector.inject(message.address, message.body)
            val inboxVerified = insertResult.optBoolean("inboxVerified")
            val inboxCount = insertResult.optInt("inboxCount")
            val defaultPackage = insertResult.optString("defaultSmsPackage")
                .ifBlank { environment.defaultSmsPackage }
            val defaultLabel = SmsEnvironment.resolveAppLabel(this, defaultPackage)

            SmsNotifier.show(this, message.address, message.body)
            api.ack(message.id, "delivered", config.deviceId)

            val userHint = when {
                !environment.hasDefaultSmsApp ->
                    "SMS вставлена в систему, но нет приложения SMS по умолчанию. Назначьте Google Messages."
                inboxVerified ->
                    "SMS должна быть в «Сообщениях» от $defaultLabel. Откройте диалог ${message.address}."
                inboxCount > 0 ->
                    "Запись в inbox есть, но текст не совпал. Откройте «Сообщения» и проверьте ${message.address}."
                else ->
                    "Shizuku не подтвердил запись в inbox. Проверьте разрешения и SMS-приложение."
            }

            val detail = buildString {
                appendLine("Серверу отправлено: delivered")
                appendLine("Проверка inbox: ${if (inboxVerified) "OK" else "не подтверждено"}")
                appendLine("Записей от ${message.address}: $inboxCount")
                append("Фрагмент из inbox: ${insertResult.optString("lastBodyPreview")}")
            }

            AgentDiagnostics.recordDelivery(
                this,
                AgentDiagnostics.DeliveryRecord(
                    messageId = message.id,
                    address = message.address,
                    bodyPreview = bodyPreview,
                    serverAck = "delivered",
                    inboxVerified = inboxVerified,
                    inboxCount = inboxCount,
                    defaultSmsPackage = defaultPackage,
                    defaultSmsLabel = defaultLabel,
                    userHint = userHint,
                    detail = detail,
                    timeMs = System.currentTimeMillis(),
                ),
            )
        } catch (error: Exception) {
            val text = error.message ?: "unknown"
            api.ack(message.id, "failed", config.deviceId, text)
            AgentDiagnostics.recordDelivery(
                this,
                AgentDiagnostics.DeliveryRecord(
                    messageId = message.id,
                    address = message.address,
                    bodyPreview = bodyPreview,
                    serverAck = "failed",
                    inboxVerified = false,
                    inboxCount = 0,
                    defaultSmsPackage = environment.defaultSmsPackage,
                    defaultSmsLabel = environment.defaultSmsLabel,
                    userHint = "Доставка не удалась: $text",
                    detail = text,
                    timeMs = System.currentTimeMillis(),
                ),
            )
        }
    }

    private fun createChannel() {
        val manager = getSystemService(NotificationManager::class.java)
        val channel = NotificationChannel(
            CHANNEL_ID,
            getString(R.string.agent_channel_name),
            NotificationManager.IMPORTANCE_LOW,
        )
        manager.createNotificationChannel(channel)
    }

    private fun buildNotification(text: String): Notification {
        return NotificationCompat.Builder(this, CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_calculator)
            .setContentTitle(getString(R.string.app_name))
            .setContentText(text)
            .setOngoing(true)
            .build()
    }

    private fun updateNotification(text: String) {
        val manager = getSystemService(NotificationManager::class.java)
        manager.notify(NOTIFICATION_ID, buildNotification(text))
    }

    companion object {
        private const val CHANNEL_ID = "sms_agent"
        private const val NOTIFICATION_ID = 42
        private const val POLL_INTERVAL_SECONDS = 5L

        fun ensureRunning(context: Context) {
            val intent = Intent(context, AgentService::class.java)
            context.startForegroundService(intent)
        }
    }
}
