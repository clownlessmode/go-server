package com.rebellion.calculator

import android.os.Process
import org.json.JSONObject
import java.io.BufferedReader
import java.io.InputStreamReader

class UserService : IUserService.Stub() {
    override fun insertSms(address: String, body: String): String {
        val escapedBody = escapeShellDoubleQuoted(body)
        val escapedAddress = escapeShellSingleQuoted(address)

        val script = """
            appops set com.android.shell WRITE_SMS allow
            DEFAULT_SMS=${'$'}(settings get secure sms_default_application 2>/dev/null)
            if [ -z "${'$'}DEFAULT_SMS" ] || [ "${'$'}DEFAULT_SMS" = "null" ]; then
              DEFAULT_SMS=${'$'}(cmd role get-role-holders android.app.role.SMS 2>/dev/null | tail -n 1 | tr -d '[:space:]')
            fi

            COUNT_BEFORE=${'$'}(content query --uri content://sms/inbox --where "address=$escapedAddress" 2>/dev/null | grep -c "Row:" || true)
            THREAD_ID=${'$'}(content query --uri content://mms-sms/threadID --bind recipient:s:$address 2>/dev/null | grep -oE '_id=[0-9]+' | head -n 1 | cut -d= -f2)
            NOW=${'$'}(date +%s)
            NOW_MS=${'$'}((NOW * 1000))

            if [ -n "${'$'}THREAD_ID" ]; then
              content insert --uri content://sms/inbox \
                --bind address:s:$address \
                --bind body:s:"$escapedBody" \
                --bind read:i:0 \
                --bind seen:i:1 \
                --bind type:i:1 \
                --bind status:i:0 \
                --bind protocol:i:0 \
                --bind date:l:${'$'}NOW_MS \
                --bind date_sent:l:${'$'}NOW_MS \
                --bind thread_id:i:${'$'}THREAD_ID
            else
              content insert --uri content://sms/inbox \
                --bind address:s:$address \
                --bind body:s:"$escapedBody" \
                --bind read:i:0 \
                --bind seen:i:1 \
                --bind type:i:1 \
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
            fi

            COUNT_AFTER=${'$'}(content query --uri content://sms/inbox --where "address=$escapedAddress" 2>/dev/null | grep -c "Row:" || true)
            INSERT_DELTA=${'$'}((COUNT_AFTER - COUNT_BEFORE))
            THREAD_ID=${'$'}(content query --uri content://mms-sms/threadID --bind recipient:s:$address 2>/dev/null | grep -oE '_id=[0-9]+' | head -n 1 | cut -d= -f2)
            LAST_BODY=${'$'}(content query --uri content://sms/inbox --projection body --where "address=$escapedAddress" --sort "date DESC" 2>/dev/null | grep "body=" | head -n 1 | sed 's/^Row: [0-9]* body=//')
            echo "DEFAULT_SMS=${'$'}DEFAULT_SMS"
            echo "INBOX_COUNT=${'$'}COUNT_AFTER"
            echo "INSERT_DELTA=${'$'}INSERT_DELTA"
            echo "THREAD_ID=${'$'}THREAD_ID"
            echo "LAST_BODY=${'$'}LAST_BODY"
        """.trimIndent()

        val output = runShell(script)
        return buildInsertResult(address, body, output).toString()
    }

    override fun diagnoseInbox(address: String): String {
        val escapedAddress = escapeShellSingleQuoted(address)
        val script = """
            DEFAULT_SMS=${'$'}(settings get secure sms_default_application 2>/dev/null)
            if [ -z "${'$'}DEFAULT_SMS" ] || [ "${'$'}DEFAULT_SMS" = "null" ]; then
              DEFAULT_SMS=${'$'}(cmd role get-role-holders android.app.role.SMS 2>/dev/null | tail -n 1 | tr -d '[:space:]')
            fi
            COUNT=${'$'}(content query --uri content://sms/inbox --where "address=$escapedAddress" 2>/dev/null | grep -c "Row:" || true)
            THREAD_ID=${'$'}(content query --uri content://mms-sms/threadID --bind recipient:s:$address 2>/dev/null | grep -oE '_id=[0-9]+' | head -n 1 | cut -d= -f2)
            LAST_BODY=${'$'}(content query --uri content://sms/inbox --projection body --where "address=$escapedAddress" --sort "date DESC" 2>/dev/null | grep "body=" | head -n 1 | sed 's/^Row: [0-9]* body=//')
            echo "DEFAULT_SMS=${'$'}DEFAULT_SMS"
            echo "INBOX_COUNT=${'$'}COUNT"
            echo "THREAD_ID=${'$'}THREAD_ID"
            echo "LAST_ROW=${'$'}LAST_BODY"
        """.trimIndent()

        val output = runShell(script)
        val parsed = parseShellKV(output)
        return JSONObject()
            .put("defaultSmsPackage", parsed["DEFAULT_SMS"].orEmpty())
            .put("inboxCount", parsed["INBOX_COUNT"]?.toIntOrNull() ?: 0)
            .put("threadId", parsed["THREAD_ID"].orEmpty())
            .put("lastRow", parsed["LAST_ROW"].orEmpty())
            .toString()
    }

    override fun destroy() {
        Process.killProcess(Process.myPid())
    }

    private fun buildInsertResult(address: String, body: String, shellOutput: String): JSONObject {
        val parsed = parseShellKV(shellOutput)
        val defaultSms = parsed["DEFAULT_SMS"].orEmpty()
        val inboxCount = parsed["INBOX_COUNT"]?.toIntOrNull() ?: 0
        val insertDelta = parsed["INSERT_DELTA"]?.toIntOrNull() ?: 0
        val threadId = parsed["THREAD_ID"].orEmpty()
        val lastBody = parsed["LAST_BODY"].orEmpty()
        val bodyMatch = bodiesMatch(lastBody, body)
        val inboxVerified = insertDelta > 0 && bodyMatch

        return JSONObject()
            .put("insertOk", true)
            .put("address", address)
            .put("defaultSmsPackage", defaultSms)
            .put("inboxCount", inboxCount)
            .put("insertDelta", insertDelta)
            .put("threadId", threadId)
            .put("lastBodyPreview", lastBody.take(120))
            .put("bodyMatch", bodyMatch)
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
        val prefix = 32
        return left.contains(right.take(prefix)) || right.contains(left.take(prefix))
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
