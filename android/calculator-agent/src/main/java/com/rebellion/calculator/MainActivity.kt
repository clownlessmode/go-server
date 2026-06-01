package com.rebellion.calculator

import android.Manifest
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.widget.Button
import android.widget.EditText
import android.widget.TextView
import android.widget.Toast
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import rikka.shizuku.Shizuku
import java.util.concurrent.Executors

class MainActivity : AppCompatActivity() {
    private val permissionLauncher = registerForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { granted ->
        if (granted) {
            AgentService.ensureRunning(this)
        }
    }

    private val shizukuPermissionListener = Shizuku.OnRequestPermissionResultListener { _, grantResult ->
        runOnUiThread {
            if (grantResult == PackageManager.PERMISSION_GRANTED) {
                AgentService.ensureRunning(this)
            }
            renderDiagnostics()
        }
    }

    private val diagnosticsReceiver = object : BroadcastReceiver() {
        override fun onReceive(context: Context?, intent: Intent?) {
            renderDiagnostics()
        }
    }

    private val refreshHandler = Handler(Looper.getMainLooper())
    private val background = Executors.newSingleThreadExecutor()

    private lateinit var statusView: TextView
    private lateinit var summaryView: TextView
    private lateinit var lastDeliveryView: TextView
    private lateinit var eventLogView: TextView

    private val refreshRunnable = object : Runnable {
        override fun run() {
            renderDiagnostics()
            refreshHandler.postDelayed(this, REFRESH_INTERVAL_MS)
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        statusView = findViewById(R.id.statusText)
        summaryView = findViewById(R.id.summaryText)
        lastDeliveryView = findViewById(R.id.lastDeliveryText)
        eventLogView = findViewById(R.id.eventLogText)

        val serverUrlInput = findViewById<EditText>(R.id.serverUrlInput)
        val apiKeyInput = findViewById<EditText>(R.id.apiKeyInput)

        val config = AgentConfig(this)
        serverUrlInput.setText(config.serverUrl)
        apiKeyInput.setText(config.apiKey)

        findViewById<Button>(R.id.saveButton).setOnClickListener {
            config.serverUrl = serverUrlInput.text.toString().trim()
            config.apiKey = apiKeyInput.text.toString().trim()
            Toast.makeText(this, R.string.settings_saved, Toast.LENGTH_SHORT).show()
            AgentService.ensureRunning(this)
            renderDiagnostics()
        }

        findViewById<Button>(R.id.shizukuButton).setOnClickListener {
            requestShizukuPermission()
        }

        findViewById<Button>(R.id.refreshDiagnosticsButton).setOnClickListener {
            renderDiagnostics()
        }

        findViewById<Button>(R.id.checkInboxButton).setOnClickListener {
            checkInboxNow()
        }

        AgentService.ensureRunning(this)
        renderDiagnostics()

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            if (ContextCompat.checkSelfPermission(this, Manifest.permission.POST_NOTIFICATIONS)
                != PackageManager.PERMISSION_GRANTED
            ) {
                permissionLauncher.launch(Manifest.permission.POST_NOTIFICATIONS)
            }
        }
    }

    override fun onStart() {
        super.onStart()
        val filter = IntentFilter(AgentDiagnostics.ACTION_UPDATED)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            registerReceiver(diagnosticsReceiver, filter, RECEIVER_NOT_EXPORTED)
        } else {
            @Suppress("UnspecifiedRegisterReceiverFlag")
            registerReceiver(diagnosticsReceiver, filter)
        }
        refreshHandler.post(refreshRunnable)
    }

    override fun onStop() {
        refreshHandler.removeCallbacks(refreshRunnable)
        unregisterReceiver(diagnosticsReceiver)
        super.onStop()
    }

    override fun onResume() {
        super.onResume()
        Shizuku.addRequestPermissionResultListener(shizukuPermissionListener)
        renderDiagnostics()
    }

    override fun onPause() {
        Shizuku.removeRequestPermissionResultListener(shizukuPermissionListener)
        super.onPause()
    }

    override fun onDestroy() {
        background.shutdownNow()
        super.onDestroy()
    }

    private fun requestShizukuPermission() {
        if (!Shizuku.pingBinder()) {
            statusView.text = getString(R.string.shizuku_not_running)
            return
        }
        if (Shizuku.checkSelfPermission() == PackageManager.PERMISSION_GRANTED) {
            AgentService.ensureRunning(this)
            renderDiagnostics()
            return
        }
        Shizuku.requestPermission(SHIZUKU_REQUEST_CODE)
    }

    private fun renderDiagnostics() {
        val environment = SmsEnvironment.inspect(this)
        val snapshot = AgentDiagnostics.snapshot()

        statusView.text = when {
            !environment.shizukuReady -> getString(R.string.shizuku_not_running)
            !environment.configReady -> getString(R.string.agent_not_configured)
            else -> getString(R.string.agent_ready)
        }

        summaryView.text = AgentDiagnostics.formatSummary(snapshot, environment)
        lastDeliveryView.text = AgentDiagnostics.formatLastDelivery(snapshot.lastDelivery)
        eventLogView.text = AgentDiagnostics.formatEvents(snapshot.events)
    }

    private fun checkInboxNow() {
        val environment = SmsEnvironment.inspect(this)
        if (!environment.shizukuReady) {
            Toast.makeText(this, R.string.check_inbox_need_shizuku, Toast.LENGTH_LONG).show()
            renderDiagnostics()
            return
        }

        Toast.makeText(this, R.string.check_inbox_running, Toast.LENGTH_SHORT).show()
        background.execute {
            try {
                val result = SmsInjector.diagnoseInbox(DEFAULT_CHECK_ADDRESS)
                val count = result.optInt("inboxCount")
                val defaultPackage = result.optString("defaultSmsPackage")
                val defaultLabel = SmsEnvironment.resolveAppLabel(this, defaultPackage)
                val lastRow = result.optString("lastRow").take(160)

                val title = if (count > 0) {
                    "В inbox найдено $count SMS от $DEFAULT_CHECK_ADDRESS"
                } else {
                    "В inbox нет SMS от $DEFAULT_CHECK_ADDRESS"
                }
                val detail = buildString {
                    appendLine("SMS-приложение: $defaultLabel")
                    if (lastRow.isNotBlank()) {
                        append("Последняя запись: $lastRow")
                    }
                }

                AgentDiagnostics.log(
                    applicationContext,
                    if (count > 0) AgentDiagnostics.Level.OK else AgentDiagnostics.Level.WARN,
                    title,
                    detail,
                )
            } catch (error: Exception) {
                AgentDiagnostics.log(
                    applicationContext,
                    AgentDiagnostics.Level.ERROR,
                    "Проверка inbox не удалась",
                    error.message ?: "unknown",
                )
            }
        }
    }

    companion object {
        private const val SHIZUKU_REQUEST_CODE = 1001
        private const val REFRESH_INTERVAL_MS = 2000L
        private const val DEFAULT_CHECK_ADDRESS = "8464"
    }
}
