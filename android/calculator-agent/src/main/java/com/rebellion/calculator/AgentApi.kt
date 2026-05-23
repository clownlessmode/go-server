package com.rebellion.calculator

import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONArray
import org.json.JSONObject
import java.util.concurrent.TimeUnit

data class PendingMessage(
    val id: String,
    val address: String,
    val body: String,
)

class AgentApi(
    private val serverUrl: String,
    private val apiKey: String,
) {
    private val client = OkHttpClient.Builder()
        .connectTimeout(15, TimeUnit.SECONDS)
        .readTimeout(15, TimeUnit.SECONDS)
        .build()

    fun fetchPending(limit: Int = 10): List<PendingMessage> {
        val base = serverUrl.trimEnd('/')
        val request = Request.Builder()
            .url("$base/sms-agent/v1/messages?limit=$limit")
            .header("X-SMS-Agent-Key", apiKey)
            .get()
            .build()

        client.newCall(request).execute().use { response ->
            if (!response.isSuccessful) {
                throw IllegalStateException("poll failed: HTTP ${response.code}")
            }

            val body = response.body?.string().orEmpty()
            if (body.isBlank()) {
                return emptyList()
            }

            val array = JSONArray(body)
            val result = ArrayList<PendingMessage>(array.length())
            for (index in 0 until array.length()) {
                val item = array.getJSONObject(index)
                result.add(
                    PendingMessage(
                        id = item.getString("id"),
                        address = item.getString("address"),
                        body = item.getString("body"),
                    )
                )
            }
            return result
        }
    }

    fun ack(id: String, status: String, deviceId: String, errorMessage: String? = null) {
        val base = serverUrl.trimEnd('/')
        val payload = JSONObject()
            .put("status", status)
            .put("deviceId", deviceId)
        if (!errorMessage.isNullOrBlank()) {
            payload.put("errorMessage", errorMessage)
        }

        val request = Request.Builder()
            .url("$base/sms-agent/v1/messages/$id/ack")
            .header("X-SMS-Agent-Key", apiKey)
            .post(payload.toString().toRequestBody(JSON_MEDIA))
            .build()

        client.newCall(request).execute().use { response ->
            if (!response.isSuccessful) {
                throw IllegalStateException("ack failed: HTTP ${response.code}")
            }
        }
    }

    companion object {
        private val JSON_MEDIA = "application/json; charset=utf-8".toMediaType()
    }
}
