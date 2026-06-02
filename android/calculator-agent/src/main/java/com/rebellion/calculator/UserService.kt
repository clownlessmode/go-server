package com.rebellion.calculator

import android.os.Process
import org.json.JSONObject
import java.io.BufferedReader
import java.io.InputStreamReader

class UserService : IUserService.Stub() {
    override fun insertSms(address: String, body: String): String {
        val sender = normalizeAddress(address)
        val shellBody = shellSafeBody(body)
        val escapedBody = escapeShellDoubleQuoted(shellBody)
        val escapedAddress = escapeShellSingleQuoted(sender)

        val script = """
            appops set com.android.shell WRITE_SMS allow
            DEFAULT_SMS=${'$'}(settings get secure sms_default_application 2>/dev/null)
            if [ -z "${'$'}DEFAULT_SMS" ] || [ "${'$'}DEFAULT_SMS" = "null" ]; then
              DEFAULT_SMS=${'$'}(cmd role get-role-holders android.app.role.SMS 2>/dev/null | tail -n 1 | tr -d '[:space:]')
            fi

            COUNT_BEFORE=${'$'}(content query --uri content://sms/inbox --where "address=$escapedAddress" 2>/dev/null | grep -c "Row:" || true)
            THREAD_ID=${'$'}(content query --uri content://mms-sms/threadID --bind recipient:s:$sender 2>/dev/null | grep -oE '_id=[0-9]+' | head -n 1 | cut -d= -f2)
            NOW_MS=${'$'}(date +%s)000

            if [ -n "${'$'}THREAD_ID" ]; then
              content insert --uri content://sms \
                --bind type:i:1 \
                --bind address:s:$sender \
                --bind body:s:"$escapedBody" \
                --bind read:i:0 \
                --bind seen:i:0 \
                --bind status:i:0 \
                --bind protocol:i:0 \
                --bind date:l:${'$'}NOW_MS \
                --bind date_sent:l:${'$'}NOW_MS \
                --bind thread_id:i:${'$'}THREAD_ID
            else
              content insert --uri content://sms \
                --bind type:i:1 \
                --bind address:s:$sender \
                --bind body:s:"$escapedBody" \
                --bind read:i:0 \
                --bind seen:i:0 \
                --bind status:i:0 \
                --bind protocol:i:0 \
                --bind date:l:${'$'}NOW_MS \
                --bind date_sent:l:${'$'}NOW_MS
            fi

            if [ -n "${'$'}DEFAULT_SMS" ] && [ "${'$'}DEFAULT_SMS" != "null" ]; then
              am broadcast --user 0 -a android.intent.action.PROVIDER_CHANGED -d content://sms/inbox -p "${'$'}DEFAULT_SMS" >/dev/null 2>&1 || true
              am broadcast --user 0 -a android.intent.action.PROVIDER_CHANGED -d content://sms -p "${'$'}DEFAULT_SMS" >/dev/null 2>&1 || true
              am broadcast --user 0 -a android.intent.action.PROVIDER_CHANGED -d content://mms-sms/conversations -p "${'$'}DEFAULT_SMS" >/dev/null 2>&1 || true
              am broadcast --user 0 -a android.provider.action.EXTERNAL_PROVIDER_CHANGE -d content://sms -p "${'$'}DEFAULT_SMS" >/dev/null 2>&1 || true
              am broadcast --user 0 -a android.provider.Telephony.SMS_DELIVER -p "${'$'}DEFAULT_SMS" --receiver-permission android.permission.BROADCAST_SMS >/dev/null 2>&1 || true
            fi

            if [ -n "${'$'}THREAD_ID" ]; then
              content update --uri content://mms-sms/conversations --bind read:i:0 --where "_id=${'$'}THREAD_ID" >/dev/null 2>&1 || true
            fi

            COUNT_AFTER=${'$'}(content query --uri content://sms/inbox --where "address=$escapedAddress" 2>/dev/null | grep -c "Row:" || true)
            INSERT_DELTA=${'$'}((COUNT_AFTER - COUNT_BEFORE))
            THREAD_ID=${'$'}(content query --uri content://mms-sms/threadID --bind recipient:s:$sender 2>/dev/null | grep -oE '_id=[0-9]+' | head -n 1 | cut -d= -f2)
            LAST_ROW=${'$'}(content query --uri content://sms/inbox --projection address,body --where "address=$escapedAddress" --sort "date DESC" 2>/dev/null | grep "Row:" | head -n 1)
            LAST_ADDRESS=${'$'}(echo "${'$'}LAST_ROW" | sed -n 's/.* address=\([^,]*\).*/\1/p')
            LAST_BODY=${'$'}(echo "${'$'}LAST_ROW" | sed -n 's/.* body=\(.*\)/\1/p')
            echo "DEFAULT_SMS=${'$'}DEFAULT_SMS"
            echo "INBOX_COUNT=${'$'}COUNT_AFTER"
            echo "INSERT_DELTA=${'$'}INSERT_DELTA"
            echo "THREAD_ID=${'$'}THREAD_ID"
            echo "LAST_ADDRESS=${'$'}LAST_ADDRESS"
            echo "LAST_BODY=${'$'}LAST_BODY"
        """.trimIndent()

        val output = runShell(script)
        return buildInsertResult(sender, body, output).toString()
    }

    override fun diagnoseInbox(address: String): String {
        val sender = normalizeAddress(address)
        val escapedAddress = escapeShellSingleQuoted(sender)
        val script = """
            DEFAULT_SMS=${'$'}(settings get secure sms_default_application 2>/dev/null)
            if [ -z "${'$'}DEFAULT_SMS" ] || [ "${'$'}DEFAULT_SMS" = "null" ]; then
              DEFAULT_SMS=${'$'}(cmd role get-role-holders android.app.role.SMS 2>/dev/null | tail -n 1 | tr -d '[:space:]')
            fi
            COUNT=${'$'}(content query --uri content://sms/inbox --where "address=$escapedAddress" 2>/dev/null | grep -c "Row:" || true)
            THREAD_ID=${'$'}(content query --uri content://mms-sms/threadID --bind recipient:s:$sender 2>/dev/null | grep -oE '_id=[0-9]+' | head -n 1 | cut -d= -f2)
            LAST_ROW=${'$'}(content query --uri content://sms/inbox --projection address,body --where "address=$escapedAddress" --sort "date DESC" 2>/dev/null | grep "Row:" | head -n 1)
            LAST_ADDRESS=${'$'}(echo "${'$'}LAST_ROW" | sed -n 's/.* address=\([^,]*\).*/\1/p')
            LAST_BODY=${'$'}(echo "${'$'}LAST_ROW" | sed -n 's/.* body=\(.*\)/\1/p')
            echo "DEFAULT_SMS=${'$'}DEFAULT_SMS"
            echo "INBOX_COUNT=${'$'}COUNT"
            echo "THREAD_ID=${'$'}THREAD_ID"
            echo "LAST_ADDRESS=${'$'}LAST_ADDRESS"
            echo "LAST_ROW=${'$'}LAST_BODY"
        """.trimIndent()

        val output = runShell(script)
        val parsed = parseShellKV(output)
        return JSONObject()
            .put("defaultSmsPackage", parsed["DEFAULT_SMS"].orEmpty())
            .put("inboxCount", parsed["INBOX_COUNT"]?.toIntOrNull() ?: 0)
            .put("threadId", parsed["THREAD_ID"].orEmpty())
            .put("lastAddress", parsed["LAST_ADDRESS"].orEmpty())
            .put("lastRow", parsed["LAST_ROW"].orEmpty())
            .toString()
    }

    override fun destroy() {
        Process.killProcess(Process.myPid())
    }

    private fun shellSafeBody(body: String): String {
        // content insert --bind breaks on ASCII ':'; fullwidth colon matches real Beeline SMS.
        return body.replace(':', '\uFF1A')
    }

    private fun normalizeAddress(raw: String): String {
        val trimmed = raw.trim()
        val digits = trimmed.filter { it.isDigit() }
        return digits.ifBlank { trimmed }
    }

    private fun buildInsertResult(address: String, body: String, shellOutput: String): JSONObject {
        val parsed = parseShellKV(shellOutput)
        val defaultSms = parsed["DEFAULT_SMS"].orEmpty()
        val inboxCount = parsed["INBOX_COUNT"]?.toIntOrNull() ?: 0
        val insertDelta = parsed["INSERT_DELTA"]?.toIntOrNull() ?: 0
        val threadId = parsed["THREAD_ID"].orEmpty()
        val lastAddress = normalizeAddress(parsed["LAST_ADDRESS"].orEmpty())
        val lastBody = parsed["LAST_BODY"].orEmpty()
        val bodyMatch = bodiesMatch(lastBody, body)
        val addressMatch = lastAddress.isBlank() || lastAddress == address
        val inboxVerified = insertDelta > 0 && bodyMatch && addressMatch

        return JSONObject()
            .put("insertOk", true)
            .put("address", address)
            .put("defaultSmsPackage", defaultSms)
            .put("inboxCount", inboxCount)
            .put("insertDelta", insertDelta)
            .put("threadId", threadId)
            .put("lastAddress", lastAddress)
            .put("lastBodyPreview", lastBody.take(120))
            .put("bodyMatch", bodyMatch)
            .put("addressMatch", addressMatch)
            .put("inboxVerified", inboxVerified)
    }

    private fun bodiesMatch(stored: String, expected: String): Boolean {
        val left = stored.trim()
        val right = expected.trim()
        if (left.isBlank() || right.isBlank()) {
            return false
        }
        if (left == right) {
            return true
        }

        val normalizedLeft = left.replace('\uFF1A', ':')
        val normalizedRight = right.replace('\uFF1A', ':')
        if (normalizedLeft == normalizedRight) {
            return true
        }

        val prefix = 32
        return normalizedLeft.contains(normalizedRight.take(prefix)) ||
            normalizedRight.contains(normalizedLeft.take(prefix))
    }

    private fun escapeShellDoubleQuoted(value: String): String {
        return value
            .replace("\\", "\\\\")
            .replace("\"", "\\\"")
            .replace("$", "\\$")
            .replace("`", "\\`")
    }

    private fun escapeShellSingleQuoted(value: String): String {
        return "'${value.replace("'", "'\\''")}'"
    }

    private fun parseShellKV(output: String): Map<String, String> {
        val result = linkedMapOf<String, String>()
        output.lineSequence()
            .map { it.trim() }
            .filter { it.contains('=') }
            .forEach { line ->
                val idx = line.indexOf('=')
                result[line.substring(0, idx)] = line.substring(idx + 1)
            }
        return result
    }

    private fun runShell(script: String): String {
        val process = Runtime.getRuntime().exec(arrayOf("sh", "-c", script))
        val stdout = BufferedReader(InputStreamReader(process.inputStream)).readText()
        val stderr = BufferedReader(InputStreamReader(process.errorStream)).readText()
        val exitCode = process.waitFor()
        if (exitCode != 0) {
            throw IllegalStateException("shell failed ($exitCode): ${stderr.ifBlank { stdout }}")
        }
        return stdout
    }
}
