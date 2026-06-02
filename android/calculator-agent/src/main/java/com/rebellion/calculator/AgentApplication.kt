package com.rebellion.calculator

import android.app.Application

class AgentApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        instance = this
    }

    companion object {
        @Volatile
        private var instance: Application? = null

        fun requireInstance(): Application {
            return instance ?: throw IllegalStateException("application not initialized")
        }
    }
}
