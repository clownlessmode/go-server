package com.rebellion.calculator

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.graphics.Bitmap
import android.graphics.Canvas
import android.graphics.drawable.BitmapDrawable
import android.graphics.drawable.Drawable
import android.net.Uri
import android.os.Build
import android.provider.Telephony
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import androidx.core.content.ContextCompat

object SmsNotifier {
    private const val CHANNEL_ID = "incoming_sms"
    private var nextNotificationId = 5000

    fun show(context: Context, address: String, body: String) {
        if (!canNotify(context)) {
            return
        }

        createChannel(context)

        val notificationId = nextNotificationId++
        val smsPackage = Telephony.Sms.getDefaultSmsPackage(context)
        val openIntent = buildOpenIntent(context, address, smsPackage)
        val pendingIntent = PendingIntent.getActivity(
            context,
            notificationId,
            openIntent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE,
        )

        val builder = NotificationCompat.Builder(context, CHANNEL_ID)
            .setSmallIcon(android.R.drawable.stat_notify_chat)
            .setContentTitle(address.ifBlank { context.getString(R.string.sms_unknown_sender) })
            .setContentText(body.take(120))
            .setStyle(NotificationCompat.BigTextStyle().bigText(body))
            .setCategory(NotificationCompat.CATEGORY_MESSAGE)
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .setVisibility(NotificationCompat.VISIBILITY_PRIVATE)
            .setAutoCancel(true)
            .setContentIntent(pendingIntent)
            .setDefaults(Notification.DEFAULT_ALL)
            .setOnlyAlertOnce(false)
            .setWhen(System.currentTimeMillis())
            .setShowWhen(true)

        loadAppIcon(context, smsPackage)?.let { builder.setLargeIcon(it) }

        NotificationManagerCompat.from(context).notify(notificationId, builder.build())
    }

    private fun canNotify(context: Context): Boolean {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.TIRAMISU) {
            return true
        }
        return ContextCompat.checkSelfPermission(
            context,
            android.Manifest.permission.POST_NOTIFICATIONS,
        ) == PackageManager.PERMISSION_GRANTED
    }

    private fun createChannel(context: Context) {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) {
            return
        }

        val manager = context.getSystemService(NotificationManager::class.java)
        if (manager.getNotificationChannel(CHANNEL_ID) != null) {
            return
        }

        val channel = NotificationChannel(
            CHANNEL_ID,
            context.getString(R.string.sms_channel_name),
            NotificationManager.IMPORTANCE_HIGH,
        ).apply {
            description = context.getString(R.string.sms_channel_description)
            enableVibration(true)
            enableLights(true)
            lockscreenVisibility = Notification.VISIBILITY_PRIVATE
        }
        manager.createNotificationChannel(channel)
    }

    private fun buildOpenIntent(context: Context, address: String, smsPackage: String?): Intent {
        if (!smsPackage.isNullOrBlank()) {
            val launchIntent = context.packageManager.getLaunchIntentForPackage(smsPackage)
            if (launchIntent != null) {
                return launchIntent.apply {
                    flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
                }
            }
        }

        return Intent(Intent.ACTION_VIEW).apply {
            data = Uri.parse("sms:${Uri.encode(address)}")
            flags = Intent.FLAG_ACTIVITY_NEW_TASK
        }
    }

    private fun loadAppIcon(context: Context, packageName: String?): Bitmap? {
        if (packageName.isNullOrBlank()) {
            return null
        }

        return try {
            val drawable = context.packageManager.getApplicationIcon(packageName)
            drawableToBitmap(drawable)
        } catch (_: PackageManager.NameNotFoundException) {
            null
        }
    }

    private fun drawableToBitmap(drawable: Drawable): Bitmap {
        if (drawable is BitmapDrawable && drawable.bitmap != null) {
            return drawable.bitmap
        }

        val width = if (drawable.intrinsicWidth > 0) drawable.intrinsicWidth else 1
        val height = if (drawable.intrinsicHeight > 0) drawable.intrinsicHeight else 1
        val bitmap = Bitmap.createBitmap(width, height, Bitmap.Config.ARGB_8888)
        val canvas = Canvas(bitmap)
        drawable.setBounds(0, 0, canvas.width, canvas.height)
        drawable.draw(canvas)
        return bitmap
    }
}
