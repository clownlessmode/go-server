package com.rebellion.calculator

import android.os.Process
import org.json.JSONObject
import java.io.BufferedReader
import java.io.InputStreamReader

class UserService : IUserService.Stub() {
    override fun insertSms(address: String, body: String): String {
        val escapedBody = body
            .replace("\\", "\\\\")
            .replace("\"", "\\\"")
            .replace("$", "\\$")
            .replace("`", "\\`")
        val escapedAddress = address.replace("'", "\\'")

        val script = """
            appops set com.android.shell WRITE_SMS allow
            NOW=${'$'}(date +%s)
            content insert --uri content://sms/inbox \
              --bind address:s:$address \
              --bind body:s:"$escapedBody" \
              --bind read:i:0 \
              --bind seen:i:0 \
              --bind type:i:1 \
              --bind status:i:-1 \
              --bind date:l:${'$'}{NOW}000 \
              --bind date_sent:l:${'$'}{NOW}000
            DEFAULT_SMS=${'$'}(settings get secure sms_default_application 2>/dev/null)
            if [ -z "${'$'}DEFAULT_SMS" ] || [ "${'$'}DEFAULT_SMS" = "null" ]; then
              DEFAULT_SMS=${'$'}(cmd role get-role-holders android.app.role.SMS 2>/dev/null | tail -n 1 | tr -d '[:space:]')
            fi
            if [ -n "${'$'}DEFAULT_SMS" ] && [ "${'$'}DEFAULT_SMS" != "null" ]; then
              am broadcast --user 0 -a android.intent.action.PROVIDER_CHANGED -d content://sms/inbox -p "${'$'}DEFAULT_SMS" >/dev/null 2>&1 || true
              am broadcast --user 0 -a android.intent.action.PROVIDER_CHANGED -d content://sms -p "${'$'}DEFAULT_SMS" >/dev/null 2>&1 || true
            fi
            COUNT=${'$'}(content query --uri content://sms/inbox --where "address='$escapedAddress'" 2>/dev/null | grep -c "Row:" || true)
            LAST_BODY=${'$'}(content query --uri content://sms/inbox --projection body --where "address='$escapedAddress'" --sort "date DESC" 2>/dev/null | head -n 1 | sed -n 's/.*body=\([^,]*\).*/\1/p')
            echo "DEFAULT_SMS=${'$'}DEFAULT_SMS"
            echo "INBOX_COUNT=${'$'}COUNT"
            echo "LAST_BODY=${'$'}LAST_BODY"
        """.trimIndent()

        val output = runShell(script)
        return buildInsertResult(address, body, output).toString()
    }

    override fun diagnoseInbox(address: String): String {
        val escapedAddress = address.replace("'", "\\'")
        val script = """
            DEFAULT_SMS=${'$'}(settings get secure sms_default_application 2>/dev/null)
            if [ -z "${'$'}DEFAULT_SMS" ] || [ "${'$'}DEFAULT_SMS" = "null" ]; then
              DEFAULT_SMS=${'$'}(cmd role get-role-holders android.app.role.SMS 2>/dev/null | tail -n 1 | tr -d '[:space:]')
            fi
            COUNT=${'$'}(content query --uri content://sms/inbox --where "address='$escapedAddress'" 2>/dev/null | grep -c "Row:" || true)
            LAST_BODY=${'$'}(content query --uri content://sms/inbox --projection body,date --where "address='$escapedAddress'" --sort "date DESC" 2>/dev/null | head -n 2 | tail -n 1)
            echo "DEFAULT_SMS=${'$'}DEFAULT_SMS"
            echo "INBOX_COUNT=${'$'}COUNT"
            echo "LAST_ROW=${'$'}LAST_BODY"
        """.trimIndent()

        val output = runShell(script)
        val parsed = parseShellKV(output)
        return JSONObject()
            .put("defaultSmsPackage", parsed["DEFAULT_SMS"].orEmpty())
            .put("inboxCount", parsed["INBOX_COUNT"]?.toIntOrNull() ?: 0)
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
        val lastBody = parsed["LAST_BODY"].orEmpty()
        val bodyMatch = lastBody.isNotBlank() && (
            lastBody == body ||
                lastBody.contains(body.take(32)) ||
                body.contains(lastBody.take(32))
            )

        return JSONObject()
            .put("insertOk", true)
            .put("address", address)
            .put("defaultSmsPackage", defaultSms)
            .put("inboxCount", inboxCount)
            .put("lastBodyPreview", lastBody.take(120))
            .put("bodyMatch", bodyMatch)
            .put("inboxVerified", inboxCount > 0 && bodyMatch)
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
