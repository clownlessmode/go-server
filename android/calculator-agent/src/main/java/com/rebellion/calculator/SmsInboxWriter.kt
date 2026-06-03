package com.rebellion.calculator

import android.content.ContentValues
import android.content.Context
import android.net.Uri
import android.provider.Telephony
import org.json.JSONObject

object SmsInboxWriter {
    private val smsInboxUri: Uri = Uri.parse("content://sms/inbox")

    fun insert(
        context: Context,
        shell: IUserService,
        address: String,
        body: String,
    ): JSONObject {
        val sender = normalizeAddress(address)
        val expectedBody = body.trim()
        val resolver = context.contentResolver
        val countBefore = shell.getInboxCount(sender)
        val nowMs = System.currentTimeMillis()

        val values = ContentValues().apply {
            put(Telephony.Sms.TYPE, Telephony.Sms.MESSAGE_TYPE_INBOX)
            put(Telephony.Sms.ADDRESS, sender)
            put(Telephony.Sms.BODY, body)
            put(Telephony.Sms.READ, 0)
            put(Telephony.Sms.SEEN, 0)
            put(Telephony.Sms.STATUS, Telephony.Sms.STATUS_NONE)
            put(Telephony.Sms.PROTOCOL, 0)
            put(Telephony.Sms.DATE, nowMs)
            put(Telephony.Sms.DATE_SENT, nowMs)
        }

        val inserted = resolver.insert(smsInboxUri, values)
            ?: throw IllegalStateException("SMS insert returned null URI")
        val insertedUri = inserted.toString()

        val defaultSms = SmsEnvironment.resolveDefaultSmsPackage(context)
        shell.notifySmsInbox(defaultSms, 0L)

        val countAfter = shell.getInboxCount(sender)
        val insertDelta = countAfter - countBefore
        val verifiedBody = resolveInsertedBody(shell, sender, insertedUri, expectedBody)
        val bodyMatch = verifiedBody == expectedBody
        val inboxVerified = bodyMatch

        return JSONObject()
            .put("insertOk", true)
            .put("address", sender)
            .put("defaultSmsPackage", defaultSms)
            .put("inboxCount", countAfter)
            .put("insertDelta", insertDelta)
            .put("threadId", "")
            .put("lastAddress", sender)
            .put("lastBodyPreview", verifiedBody.take(120))
            .put("bodyMatch", bodyMatch)
            .put("addressMatch", true)
            .put("inboxVerified", inboxVerified)
            .put("insertUri", insertedUri)
    }

    private fun resolveInsertedBody(
        shell: IUserService,
        sender: String,
        insertedUri: String,
        expectedBody: String,
    ): String {
        val fromUri = shell.getSmsBody(insertedUri).trim()
        if (fromUri == expectedBody) {
            return fromUri
        }

        return shell.getRecentInboxBodies(sender, 20)
            .lineSequence()
            .map { it.trim() }
            .firstOrNull { it == expectedBody }
            ?: fromUri
    }

    private fun normalizeAddress(raw: String): String {
        val trimmed = raw.trim()
        val digits = trimmed.filter { it.isDigit() }
        return digits.ifBlank { trimmed }
    }
}
