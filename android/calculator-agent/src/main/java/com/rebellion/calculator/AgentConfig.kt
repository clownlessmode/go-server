package com.rebellion.calculator

import android.content.Context

class AgentConfig(context: Context) {
    private val prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)

    var serverUrl: String
        get() = prefs.getString(KEY_SERVER_URL, DEFAULT_SERVER_URL).orEmpty()
        set(value) = prefs.edit().putString(KEY_SERVER_URL, value.trim()).apply()

    var apiKey: String
        get() = prefs.getString(KEY_API_KEY, "").orEmpty()
        set(value) = prefs.edit().putString(KEY_API_KEY, value.trim()).apply()

    val deviceId: String
        get() {
            val existing = prefs.getString(KEY_DEVICE_ID, null)
            if (!existing.isNullOrBlank()) {
                return existing
            }
            val generated = java.util.UUID.randomUUID().toString()
            prefs.edit().putString(KEY_DEVICE_ID, generated).apply()
            return generated
        }

    val isConfigured: Boolean
        get() = serverUrl.isNotBlank() && apiKey.isNotBlank()

    companion object {
        private const val PREFS_NAME = "agent_config"
        private const val KEY_SERVER_URL = "server_url"
        private const val KEY_API_KEY = "api_key"
        private const val KEY_DEVICE_ID = "device_id"
        private const val DEFAULT_SERVER_URL = "http://192.168.1.100:8080"
    }
}
