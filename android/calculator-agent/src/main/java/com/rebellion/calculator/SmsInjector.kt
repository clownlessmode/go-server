package com.rebellion.calculator

import android.content.ComponentName
import android.content.ServiceConnection
import android.content.pm.PackageManager
import android.os.IBinder
import org.json.JSONObject
import rikka.shizuku.Shizuku

object SmsInjector {
    @Volatile
    private var userService: IUserService? = null

    @Volatile
    private var binding = false

    private val serviceArgs = Shizuku.UserServiceArgs(
        ComponentName(
            "com.rebellion.calculator",
            UserService::class.java.name,
        ),
    )
        .daemon(false)
        .processNameSuffix("sms_service")
        .debuggable(BuildConfig.DEBUG)
        .version(4)

    private val connection = object : ServiceConnection {
        override fun onServiceConnected(name: ComponentName?, binder: IBinder?) {
            binding = false
            if (binder == null || !binder.pingBinder()) {
                userService = null
                return
            }
            userService = IUserService.Stub.asInterface(binder)
        }

        override fun onServiceDisconnected(name: ComponentName?) {
            userService = null
        }
    }

    private val binderListener = Shizuku.OnBinderReceivedListener {
        bindService()
    }

    private val deathListener = Shizuku.OnBinderDeadListener {
        userService = null
        bindService()
    }

    fun ensureBound() {
        Shizuku.addBinderReceivedListenerSticky(binderListener)
        Shizuku.addBinderDeadListener(deathListener)
        if (Shizuku.pingBinder()) {
            bindService()
        }
    }

    fun inject(address: String, body: String): JSONObject {
        val service = requireService()
        val raw = service.insertSms(address, body)
        return AgentDiagnostics.parseInsertResult(raw)
    }

    fun diagnoseInbox(address: String): JSONObject {
        val service = requireService()
        val raw = service.diagnoseInbox(address)
        return JSONObject(raw)
    }

    private fun requireService(): IUserService {
        if (!Shizuku.pingBinder()) {
            throw IllegalStateException("Shizuku is not running")
        }
        if (Shizuku.checkSelfPermission() != PackageManager.PERMISSION_GRANTED) {
            throw IllegalStateException("Shizuku permission not granted")
        }

        ensureBound()
        waitForService()

        return userService ?: throw IllegalStateException("Shizuku user service unavailable")
    }

    private fun bindService() {
        if (userService != null || binding) {
            return
        }
        binding = true
        Shizuku.bindUserService(serviceArgs, connection)
    }

    private fun waitForService() {
        if (userService != null) {
            return
        }
        bindService()
        repeat(20) {
            if (userService != null) {
                return
            }
            Thread.sleep(100)
        }
    }
}
