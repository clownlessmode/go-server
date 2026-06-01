package com.rebellion.calculator

import android.content.Context
import android.content.pm.PackageManager
import android.provider.Telephony
import rikka.shizuku.Shizuku

object SmsEnvironment {
    data class Info(
        val shizukuStatus: String,
        val shizukuReady: Boolean,
        val configStatus: String,
        val configReady: Boolean,
        val defaultSmsPackage: String,
        val defaultSmsLabel: String,
        val hasDefaultSmsApp: Boolean,
    )

    fun inspect(context: Context): Info {
        val config = AgentConfig(context)
        val defaultPackage = resolveDefaultSmsPackage(context)
        val defaultLabel = resolveAppLabel(context, defaultPackage)

        val shizukuReady = Shizuku.pingBinder() &&
            Shizuku.checkSelfPermission() == PackageManager.PERMISSION_GRANTED
        val shizukuStatus = when {
            !Shizuku.pingBinder() -> "не запущен — откройте «Блокнот» (Shizuku)"
            Shizuku.checkSelfPermission() != PackageManager.PERMISSION_GRANTED ->
                "нет разрешения — нажмите «Запросить Shizuku»"
            else -> "подключён"
        }

        val configReady = config.isConfigured
        val configStatus = if (configReady) {
            "OK (${config.serverUrl.trimEnd('/')})"
        } else {
            "укажите URL сервера и ключ"
        }

        val hasDefaultSms = defaultPackage.isNotBlank()
        val smsLabel = if (hasDefaultSms) {
            defaultLabel
        } else {
            "не выбрано — SMS может не появиться в «Сообщениях»"
        }

        return Info(
            shizukuStatus = shizukuStatus,
            shizukuReady = shizukuReady,
            configStatus = configStatus,
            configReady = configReady,
            defaultSmsPackage = defaultPackage,
            defaultSmsLabel = smsLabel,
            hasDefaultSmsApp = hasDefaultSms,
        )
    }

    fun resolveDefaultSmsPackage(context: Context): String {
        return Telephony.Sms.getDefaultSmsPackage(context)?.trim().orEmpty()
    }

    fun resolveAppLabel(context: Context, packageName: String): String {
        if (packageName.isBlank()) {
            return "—"
        }
        return try {
            val pm = context.packageManager
            pm.getApplicationLabel(pm.getApplicationInfo(packageName, 0)).toString()
        } catch (_: PackageManager.NameNotFoundException) {
            packageName
        }
    }
}
