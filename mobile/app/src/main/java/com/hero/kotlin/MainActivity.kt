package com.hero.kotlin

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import com.hero.kotlin.ui.theme.HeroTheme
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint  // указываем что Activity использует Hilt для dependency injection
class MainActivity : ComponentActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // Устанавливаем Compose как UI фреймворк
        setContent {
            HeroTheme {
                // Основной контейнер приложения
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    MainScreen()  // основной экран приложения
                }
            }
        }
    }
}

// Основной экран приложения
@Composable
fun MainScreen() {
    var currentScreen by remember { mutableStateOf("home") }  // текущий экран

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Top
    ) {
        // Заголовок приложения
        Text(
            text = "HERO!N Messenger",
            style = MaterialTheme.typography.headlineMedium,
            color = MaterialTheme.colorScheme.primary,
            modifier = Modifier.padding(bottom = 32.dp)
        )

        // Статус подключения
        Card(
            modifier = Modifier
                .fillMaxWidth()
                .padding(bottom = 16.dp)
        ) {
            Column(
                modifier = Modifier.padding(16.dp)
            ) {
                Text(
                    text = "Статус сети",
                    style = MaterialTheme.typography.titleMedium
                )
                Text(
                    text = "Подключение к P2P сети...",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }

        // Кнопки навигации
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceEvenly
        ) {
            Button(
                onClick = { currentScreen = "messages" },
                modifier = Modifier.weight(1f).padding(end = 8.dp)
            ) {
                Text("Сообщения")
            }

            Button(
                onClick = { currentScreen = "contacts" },
                modifier = Modifier.weight(1f).padding(start = 8.dp)
            ) {
                Text("Контакты")
            }
        }

        // Контент в зависимости от текущего экрана
        when (currentScreen) {
            "messages" -> MessagesScreen()
            "contacts" -> ContactsScreen()
            else -> HomeScreen()
        }
    }
}

// Экран сообщений
@Composable
fun MessagesScreen() {
    Column(
        modifier = Modifier.fillMaxWidth(),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(
            text = "Сообщения",
            style = MaterialTheme.typography.headlineSmall,
            modifier = Modifier.padding(bottom = 16.dp)
        )

        Text(
            text = "Здесь будут отображаться сообщения",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

// Экран контактов
@Composable
fun ContactsScreen() {
    Column(
        modifier = Modifier.fillMaxWidth(),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(
            text = "Контакты",
            style = MaterialTheme.typography.headlineSmall,
            modifier = Modifier.padding(bottom = 16.dp)
        )

        Text(
            text = "Здесь будут отображаться контакты",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

// Домашний экран
@Composable
fun HomeScreen() {
    Column(
        modifier = Modifier.fillMaxWidth(),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(
            text = "Добро пожаловать в HERO!N",
            style = MaterialTheme.typography.headlineSmall,
            modifier = Modifier.padding(bottom = 16.dp)
        )

        Text(
            text = "Незаблокируемый P2P месенджер",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.padding(bottom = 16.dp)
        )

        Card(
            modifier = Modifier.fillMaxWidth()
        ) {
            Column(
                modifier = Modifier.padding(16.dp)
            ) {
                Text(
                    text = "Функции:",
                    style = MaterialTheme.typography.titleMedium,
                    modifier = Modifier.padding(bottom = 8.dp)
                )

                Text("• E2E шифрование ChaCha20-Poly1305")
                Text("• P2P сеть без центральных серверов")
                Text("• Автономный режим через Wi-Fi")
                Text("• IPFS для распределенного хранения")
                Text("• Space-Time Proofs для надежности")
            }
        }
    }
}

// Preview для дизайнера
@Preview(showBackground = true)
@Composable
fun MainScreenPreview() {
    HeroTheme {
        MainScreen()
    }
}
