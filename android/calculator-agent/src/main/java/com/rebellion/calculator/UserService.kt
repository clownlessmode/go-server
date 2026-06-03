package com.rebellion.calculator

import android.os.Process
import org.json.JSONObject
import java.io.BufferedReader
import java.io.InputStreamReader

class UserService : IUserService.Stub() {
    override fun grantWriteSms(packageName: String) {
        val safePackage = escapeShellDoubleQuoted(packageName.trim())
        runShell(
            """
            appops set "$safePackage" WRITE_SMS allow
            """.trimIndent(),
        )
    }

    override fun getInboxCount(address: String): Int {
        val escapedAddress = escapeShellSingleQuoted(normalizeAddress(address))
        val output = runShell(
            """
            content query --uri content://sms/inbox --where "address=$escapedAddress" 2>/dev/null | grep -c "Row:" || true
            """.trimIndent(),
        ).trim()
        return output.toIntOrNull() ?: 0
    }

    override fun getLastInboxBody(address: String): String {
        return getRecentInboxBodies(address, 1)
            .lineSequence()
            .firstOrNull()
            .orEmpty()
    }

    override fun getSmsBody(uri: String): String {
        val safeUri = escapeShellDoubleQuoted(uri.trim())
        if (safeUri.isBlank()) {
            return ""
        }
        return runShell(
            """
            ROW=${'$'}(content query --uri "$safeUri" --projection body 2>/dev/null | grep "Row:" | head -n 1)
            echo "${'$'}(echo "${'$'}ROW" | sed 's/^Row: [0-9]* body=//')"
            """.trimIndent(),
        ).trim()
    }

    override fun getRecentInboxBodies(address: String, limit: Int): String {
        val escapedAddress = escapeShellSingleQuoted(normalizeAddress(address))
        val safeLimit = limit.coerceIn(1, 30)
        return runShell(
            """
            content query --uri content://sms/inbox --projection body --where "address=$escapedAddress" --sort "date DESC" 2>/dev/null | grep "Row:" | head -n $safeLimit | sed 's/^Row: [0-9]* body=//'
            """.trimIndent(),
        ).trimEnd()
    }

    override fun notifySmsInbox(defaultSmsPackage: String, threadId: Long): String {
        if (defaultSmsPackage.isBlank()) {
            return "skipped"
        }

        val escapedPackage = escapeShellDoubleQuoted(defaultSmsPackage)
        val threadUpdate = if (threadId > 0L) {
            """
            content update --uri content://mms-sms/conversations --bind read:i:0 --where "_id=$threadId" >/dev/null 2>&1 || true
            """.trimIndent()
        } else {
            ""
        }

        runShell(
            """
            am broadcast --user 0 -a android.intent.action.PROVIDER_CHANGED -d content://sms/inbox -p "$escapedPackage" >/dev/null 2>&1 || true
            am broadcast --user 0 -a android.intent.action.PROVIDER_CHANGED -d content://sms -p "$escapedPackage" >/dev/null 2>&1 || true
            am broadcast --user 0 -a android.intent.action.PROVIDER_CHANGED -d content://mms-sms/conversations -p "$escapedPackage" >/dev/null 2>&1 || true
            am broadcast --user 0 -a android.provider.action.EXTERNAL_PROVIDER_CHANGE -d content://sms -p "$escapedPackage" >/dev/null 2>&1 || true
            am broadcast --user 0 -a android.provider.Telephony.SMS_DELIVER -p "$escapedPackage" --receiver-permission android.permission.BROADCAST_SMS >/dev/null 2>&1 || true
            $threadUpdate
            """.trimIndent(),
        )
        return "ok"
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
            echo "DEFAULT_SMS=${'$'}DEFAULT_SMS"
            echo "INBOX_COUNT=${'$'}COUNT"
            echo "THREAD_ID=${'$'}THREAD_ID"
            echo "LAST_ADDRESS=$sender"
            echo "LAST_ROW=${'$'}(content query --uri content://sms/inbox --projection body --where "address=$escapedAddress" --sort "date DESC" 2>/dev/null | grep "Row:" | head -n 1 | sed 's/^Row: [0-9]* body=//')"
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

    private fun normalizeAddress(raw: String): String {
        val trimmed = raw.trim()
        val digits = trimmed.filter { it.isDigit() }
        return digits.ifBlank { trimmed }
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
