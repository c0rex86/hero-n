package com.hero.kotlin.ui.theme

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color

// Цветовая палитра для темной темы
private val DarkColorScheme = darkColorScheme(
    primary = Color(0xFF4CAF50),        // зеленый - основной цвет
    onPrimary = Color.White,            // белый текст на основном цвете
    primaryContainer = Color(0xFF2E7D32), // темный зеленый для контейнеров
    onPrimaryContainer = Color.White,

    secondary = Color(0xFF2196F3),      // синий - дополнительный цвет
    onSecondary = Color.White,
    secondaryContainer = Color(0xFF1976D2),
    onSecondaryContainer = Color.White,

    background = Color(0xFF121212),     // темный фон
    onBackground = Color.White,         // белый текст на фоне
    surface = Color(0xFF1E1E1E),        // цвет поверхности
    onSurface = Color.White,

    surfaceVariant = Color(0xFF2D2D2D), // вариант поверхности
    onSurfaceVariant = Color(0xFFE0E0E0), // светло-серый текст

    error = Color(0xFFCF6679),          // цвет ошибок
    onError = Color.Black,
)

// Цветовая палитра для светлой темы
private val LightColorScheme = lightColorScheme(
    primary = Color(0xFF2E7D32),        // темный зеленый для светлой темы
    onPrimary = Color.White,
    primaryContainer = Color(0xFFA5D6A7), // светлый зеленый для контейнеров
    onPrimaryContainer = Color(0xFF1B5E20),

    secondary = Color(0xFF1976D2),      // темный синий для светлой темы
    onSecondary = Color.White,
    secondaryContainer = Color(0xFF90CAF9),
    onSecondaryContainer = Color(0xFF0D47A1),

    background = Color(0xFFFFFBFE),     // светлый фон
    onBackground = Color(0xFF1C1B1F),   // темный текст на фоне
    surface = Color(0xFFFFFBFE),        // цвет поверхности
    onSurface = Color(0xFF1C1B1F),

    surfaceVariant = Color(0xFFE7E0EC), // вариант поверхности
    onSurfaceVariant = Color(0xFF49454F), // темный текст

    error = Color(0xFFB3261E),          // цвет ошибок
    onError = Color.White,
)

// Типография для приложения
private val AppTypography = Typography(
    // Здесь можно настроить шрифты, размеры текста и т.д.
)

// Формы (shapes) для компонентов
private val AppShapes = Shapes(
    // Здесь можно настроить скругления углов для компонентов
)

@Composable
fun HeroTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),  // автоматически определяем тему
    dynamicColor: Boolean = true,  // поддержка dynamic colors на Android 12+
    content: @Composable () -> Unit
) {
    val colorScheme = when {
        dynamicColor && darkTheme -> {
            // Используем dynamic colors для Android 12+ в темной теме
            dynamicDarkColorScheme(context = androidx.compose.ui.platform.LocalContext.current)
        }
        dynamicColor && !darkTheme -> {
            // Используем dynamic colors для Android 12+ в светлой теме
            dynamicLightColorScheme(context = androidx.compose.ui.platform.LocalContext.current)
        }
        darkTheme -> DarkColorScheme  // наша кастомная темная тема
        else -> LightColorScheme       // наша кастомная светлая тема
    }

    MaterialTheme(
        colorScheme = colorScheme,
        typography = AppTypography,
        shapes = AppShapes,
        content = content
    )
}
