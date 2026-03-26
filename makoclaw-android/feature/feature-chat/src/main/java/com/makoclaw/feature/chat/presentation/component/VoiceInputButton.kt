package com.makoclaw.feature.chat.presentation.component

import android.Manifest
import android.content.pm.PackageManager
import android.media.MediaRecorder
import android.os.Build
import androidx.compose.animation.animateColorAsState
import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Mic
import androidx.compose.material.icons.filled.Stop
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.IconButtonDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import java.io.File

/**
 * Voice input button that records audio for transcription.
 * Records to a temp file, then calls onRecordingComplete with the file path.
 */
@Composable
fun VoiceInputButton(
    isRecording: Boolean,
    onStartRecording: () -> Unit,
    onStopRecording: () -> Unit,
    modifier: Modifier = Modifier
) {
    val context = LocalContext.current
    val hasPermission = remember {
        ContextCompat.checkSelfPermission(
            context, Manifest.permission.RECORD_AUDIO
        ) == PackageManager.PERMISSION_GRANTED
    }

    val containerColor by animateColorAsState(
        targetValue = if (isRecording)
            MaterialTheme.colorScheme.error
        else
            MaterialTheme.colorScheme.surfaceContainer,
        label = "micColor"
    )

    IconButton(
        onClick = {
            if (!hasPermission) return@IconButton
            if (isRecording) onStopRecording() else onStartRecording()
        },
        modifier = modifier,
        colors = IconButtonDefaults.iconButtonColors(containerColor = containerColor)
    ) {
        Icon(
            imageVector = if (isRecording) Icons.Filled.Stop else Icons.Filled.Mic,
            contentDescription = if (isRecording) "Stop recording" else "Start recording",
            modifier = Modifier.size(24.dp),
            tint = if (isRecording)
                MaterialTheme.colorScheme.onError
            else
                MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

/**
 * Helper to manage audio recording to a temp file.
 */
class AudioRecorder(private val context: android.content.Context) {
    private var recorder: MediaRecorder? = null
    private var outputFile: File? = null

    fun startRecording(): File? {
        val file = File(context.cacheDir, "voice_${System.currentTimeMillis()}.m4a")
        outputFile = file

        recorder = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            MediaRecorder(context)
        } else {
            @Suppress("DEPRECATION")
            MediaRecorder()
        }.apply {
            setAudioSource(MediaRecorder.AudioSource.MIC)
            setOutputFormat(MediaRecorder.OutputFormat.MPEG_4)
            setAudioEncoder(MediaRecorder.AudioEncoder.AAC)
            setAudioEncodingBitRate(128000)
            setAudioSamplingRate(44100)
            setOutputFile(file.absolutePath)
            prepare()
            start()
        }

        return file
    }

    fun stopRecording(): File? {
        try {
            recorder?.apply {
                stop()
                release()
            }
        } catch (_: Exception) { }
        recorder = null
        return outputFile
    }

    fun release() {
        try {
            recorder?.release()
        } catch (_: Exception) { }
        recorder = null
    }
}
