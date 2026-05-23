package com.rebellion.calculator

import android.content.ComponentName
import android.content.ServiceConnection
import android.content.pm.PackageManager
import android.os.IBinder
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
        .version(1)

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

    fun inject(address: String, body: String) {
        if (!Shizuku.pingBinder()) {
            throw IllegalStateException("Shizuku is not running")
        }
        if (Shizuku.checkSelfPermission() != PackageManager.PERMISSION_GRANTED) {
            throw IllegalStateException("Shizuku permission not granted")
        }

        ensureBound()
        waitForService()

        val service = userService ?: throw IllegalStateException("Shizuku user service unavailable")
        service.insertSms(address, body)
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
