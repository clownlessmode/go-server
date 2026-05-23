package com.rebellion.calculator

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.widget.Button
import android.widget.EditText
import android.widget.TextView
import android.widget.Toast
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import androidx.core.content.ContextCompat
import rikka.shizuku.Shizuku

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
                statusView.text = getString(R.string.shizuku_granted)
                AgentService.ensureRunning(this)
            } else {
                statusView.text = getString(R.string.shizuku_denied)
            }
        }
    }

    private lateinit var statusView: TextView

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        statusView = findViewById(R.id.statusText)
        val serverUrlInput = findViewById<EditText>(R.id.serverUrlInput)
        val apiKeyInput = findViewById<EditText>(R.id.apiKeyInput)

        val config = AgentConfig(this)
        serverUrlInput.setText(config.serverUrl)
        apiKeyInput.setText(config.apiKey)

        refreshStatus()

        findViewById<Button>(R.id.saveButton).setOnClickListener {
            config.serverUrl = serverUrlInput.text.toString().trim()
            config.apiKey = apiKeyInput.text.toString().trim()
            Toast.makeText(this, R.string.settings_saved, Toast.LENGTH_SHORT).show()
            AgentService.ensureRunning(this)
        }

        findViewById<Button>(R.id.shizukuButton).setOnClickListener {
            requestShizukuPermission()
        }

        AgentService.ensureRunning(this)

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            if (ContextCompat.checkSelfPermission(this, Manifest.permission.POST_NOTIFICATIONS)
                != PackageManager.PERMISSION_GRANTED
            ) {
                permissionLauncher.launch(Manifest.permission.POST_NOTIFICATIONS)
            }
        }
    }

    override fun onResume() {
        super.onResume()
        Shizuku.addRequestPermissionResultListener(shizukuPermissionListener)
        refreshStatus()
    }

    override fun onPause() {
        Shizuku.removeRequestPermissionResultListener(shizukuPermissionListener)
        super.onPause()
    }

    private fun requestShizukuPermission() {
        if (!Shizuku.pingBinder()) {
            statusView.text = getString(R.string.shizuku_not_running)
            return
        }
        if (Shizuku.checkSelfPermission() == PackageManager.PERMISSION_GRANTED) {
            statusView.text = getString(R.string.shizuku_granted)
            AgentService.ensureRunning(this)
            return
        }
        Shizuku.requestPermission(SHIZUKU_REQUEST_CODE)
    }

    private fun refreshStatus() {
        statusView.text = when {
            !Shizuku.pingBinder() -> getString(R.string.shizuku_not_running)
            Shizuku.checkSelfPermission() != PackageManager.PERMISSION_GRANTED ->
                getString(R.string.shizuku_permission_required)
            else -> getString(R.string.agent_ready)
        }
    }

    companion object {
        private const val SHIZUKU_REQUEST_CODE = 1001
    }
}
