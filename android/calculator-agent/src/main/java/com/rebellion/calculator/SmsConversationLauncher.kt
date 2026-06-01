package com.rebellion.calculator

import android.content.Context
import android.content.Intent
import android.net.Uri

object SmsConversationLauncher {
    fun open(context: Context, address: String) {
        val normalized = address.trim()
        if (normalized.isEmpty()) {
            return
        }

        val uri = Uri.parse("sms:${Uri.encode(normalized)}")
        val intent = Intent(Intent.ACTION_VIEW, uri).apply {
            addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
        }

        val defaultSms = SmsEnvironment.resolveDefaultSmsPackage(context)
        if (defaultSms.isNotBlank()) {
            intent.setPackage(defaultSms)
        }

        try {
            context.startActivity(intent)
        } catch (_: Exception) {
            context.startActivity(
                Intent(Intent.ACTION_VIEW, uri).addFlags(Intent.FLAG_ACTIVITY_NEW_TASK),
            )
        }
    }
}
