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
        executor.scheduleWithFixedDelay({ pollOnce() }, 0, POLL_INTERVAL_SECONDS, TimeUnit.SECONDS)
    }

    override fun onDestroy() {
        executor.shutdownNow()
        super.onDestroy()
    }

    override fun onBind(intent: Intent?): IBinder? = null

    private fun pollOnce() {
        val config = AgentConfig(this)
        if (!config.isConfigured) {
            updateNotification(getString(R.string.agent_not_configured))
            return
        }
        if (!Shizuku.pingBinder() || Shizuku.checkSelfPermission() != PackageManager.PERMISSION_GRANTED) {
            updateNotification(getString(R.string.agent_waiting_shizuku))
            return
        }

        try {
            val api = AgentApi(config.serverUrl, config.apiKey)
            val messages = api.fetchPending()
            if (messages.isEmpty()) {
                updateNotification(getString(R.string.agent_idle))
                return
            }

            for (message in messages) {
                try {
                    SmsInjector.inject(message.address, message.body)
                    SmsNotifier.show(this, message.address, message.body)
                    api.ack(message.id, "delivered", config.deviceId)
                } catch (error: Exception) {
                    api.ack(message.id, "failed", config.deviceId, error.message)
                }
            }
            updateNotification(getString(R.string.agent_delivered, messages.size))
        } catch (error: Exception) {
            updateNotification(getString(R.string.agent_error, error.message ?: "unknown"))
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
