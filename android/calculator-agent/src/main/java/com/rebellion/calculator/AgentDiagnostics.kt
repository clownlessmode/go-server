package com.rebellion.calculator

import android.content.Context
import android.content.Intent
import org.json.JSONObject
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import java.util.concurrent.CopyOnWriteArrayList

object AgentDiagnostics {
    const val ACTION_UPDATED = "com.rebellion.calculator.DIAGNOSTICS_UPDATED"
    const val MAX_EVENTS = 80

    enum class Level { OK, WARN, ERROR, INFO }

    data class Event(
        val timeMs: Long,
        val level: Level,
        val title: String,
        val detail: String,
    )

    data class DeliveryRecord(
        val messageId: String,
        val address: String,
        val bodyPreview: String,
        val serverAck: String,
        val inboxVerified: Boolean,
        val inboxCount: Int,
        val insertDelta: Int,
        val threadId: String,
        val defaultSmsPackage: String,
        val defaultSmsLabel: String,
        val userHint: String,
        val detail: String,
        val timeMs: Long,
    )

    data class Snapshot(
        val events: List<Event>,
        val lastDelivery: DeliveryRecord?,
        val lastPollAt: Long?,
        val lastPollSummary: String,
        val agentState: String,
    )

    private val events = CopyOnWriteArrayList<Event>()
    @Volatile private var lastDelivery: DeliveryRecord? = null
    @Volatile private var lastPollAt: Long? = null
    @Volatile private var lastPollSummary: String = "Ещё не опрашивали сервер"
    @Volatile private var agentState: String = "Запуск…"

    private val timeFormat = SimpleDateFormat("HH:mm:ss", Locale.getDefault())

    fun setAgentState(context: Context, state: String) {
        agentState = state
        notifyUpdated(context)
    }

    fun setPollResult(context: Context, summary: String) {
        lastPollAt = System.currentTimeMillis()
        lastPollSummary = summary
        notifyUpdated(context)
    }

    fun log(context: Context, level: Level, title: String, detail: String = "") {
        events.add(0, Event(System.currentTimeMillis(), level, title, detail))
        while (events.size > MAX_EVENTS) {
            events.removeAt(events.size - 1)
        }
        notifyUpdated(context)
    }

    fun recordDelivery(context: Context, record: DeliveryRecord) {
        lastDelivery = record
        val level = when {
            record.serverAck == "failed" -> Level.ERROR
            record.inboxVerified -> Level.OK
            else -> Level.WARN
        }
        log(context, level, record.userHint, record.detail)
        notifyUpdated(context)
    }

    fun snapshot(): Snapshot {
        return Snapshot(
            events = events.toList(),
            lastDelivery = lastDelivery,
            lastPollAt = lastPollAt,
            lastPollSummary = lastPollSummary,
            agentState = agentState,
        )
    }

    fun formatEvents(events: List<Event>): String {
        if (events.isEmpty()) {
            return "Пока нет событий. Агент начнёт писать сюда после опроса сервера."
        }
        return events.joinToString(separator = "\n\n") { event ->
            val icon = when (event.level) {
                Level.OK -> "✅"
                Level.WARN -> "⚠️"
                Level.ERROR -> "❌"
                Level.INFO -> "ℹ️"
            }
            buildString {
                append(icon)
                append(' ')
                append(timeFormat.format(Date(event.timeMs)))
                append(" — ")
                append(event.title)
                if (event.detail.isNotBlank()) {
                    append('\n')
                    append(event.detail.trim())
                }
            }
        }
    }

    fun formatLastDelivery(record: DeliveryRecord?): String {
        if (record == null) {
            return "Последняя SMS ещё не доставлялась через агент."
        }
        val statusLine = when {
            record.serverAck == "failed" -> "❌ Серверу отправлен статус: ОШИБКА"
            record.inboxVerified -> "✅ SMS есть в системной базе (inbox)"
            else -> "⚠️ Вставка прошла, но SMS не найдена в inbox"
        }
        return buildString {
            appendLine(statusLine)
            appendLine("Время: ${timeFormat.format(Date(record.timeMs))}")
            appendLine("От: ${record.address}")
            appendLine("ID на сервере: ${record.messageId}")
            appendLine("Текст: ${record.bodyPreview}")
            appendLine("SMS-приложение по умолчанию: ${record.defaultSmsLabel}")
            if (record.defaultSmsPackage.isNotBlank()) {
                appendLine("Пакет: ${record.defaultSmsPackage}")
            }
            appendLine("Добавлено в inbox: ${record.insertDelta}")
            appendLine("Всего от ${record.address} в inbox: ${record.inboxCount}")
            if (record.threadId.isNotBlank()) {
                appendLine("thread_id: ${record.threadId}")
            }
            appendLine()
            appendLine("Уведомление «Калькулятора» ≠ приложение «Сообщения».")
            append(record.userHint)
            if (record.detail.isNotBlank()) {
                appendLine()
                append(record.detail.trim())
            }
        }
    }

    fun formatSummary(snapshot: Snapshot, environment: SmsEnvironment.Info): String {
        return buildString {
            appendLine("Состояние агента: ${snapshot.agentState}")
            appendLine("Shizuku: ${environment.shizukuStatus}")
            appendLine("Настройки сервера: ${environment.configStatus}")
            appendLine("SMS-приложение: ${environment.defaultSmsLabel}")
            snapshot.lastPollAt?.let {
                appendLine("Последний опрос: ${timeFormat.format(Date(it))}")
            }
            append("Очередь: ${snapshot.lastPollSummary}")
        }
    }

    fun parseInsertResult(raw: String): JSONObject {
        return try {
            JSONObject(raw)
        } catch (_: Exception) {
            JSONObject()
                .put("insertOk", false)
                .put("error", raw)
        }
    }

    private fun notifyUpdated(context: Context) {
        context.applicationContext.sendBroadcast(
            Intent(ACTION_UPDATED).setPackage(context.packageName),
        )
    }
}
