package com.rebellion.calculator

import android.os.Process
import java.io.BufferedReader
import java.io.InputStreamReader

class UserService : IUserService.Stub() {
    override fun insertSms(address: String, body: String): Int {
        val escapedBody = body
            .replace("\\", "\\\\")
            .replace("\"", "\\\"")
            .replace("$", "\\$")
            .replace("`", "\\`")

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
        """.trimIndent()

        val process = Runtime.getRuntime().exec(arrayOf("sh", "-c", script))
        val stderr = BufferedReader(InputStreamReader(process.errorStream)).readText()
        val exitCode = process.waitFor()
        if (exitCode != 0) {
            throw IllegalStateException("sms insert failed: $stderr")
        }
        return exitCode
    }

    override fun destroy() {
        Process.killProcess(Process.myPid())
    }
}
