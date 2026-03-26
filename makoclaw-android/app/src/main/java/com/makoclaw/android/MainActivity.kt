package com.makoclaw.android

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Surface
import androidx.compose.ui.Modifier
import com.makoclaw.android.navigation.MakoClawNavHost
import com.makoclaw.core.ui.theme.MakoClawTheme
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        enableEdgeToEdge()
        super.onCreate(savedInstanceState)

        setContent {
            MakoClawTheme {
                Surface(modifier = Modifier.fillMaxSize()) {
                    MakoClawNavHost()
                }
            }
        }
    }
}
