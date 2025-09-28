package dev.c0rex64.heroin.core.network

import io.grpc.ManagedChannel
import io.grpc.android.AndroidChannelBuilder
import io.grpc.okhttp.OkHttpChannelBuilder
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.util.concurrent.TimeUnit
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class GrpcClient @Inject constructor() {
    
    private var channel: ManagedChannel? = null
    private var currentEndpoint: String? = null
    
    // подключиться к серверу
    suspend fun connect(endpoint: String, useTls: Boolean = true) = withContext(Dispatchers.IO) {
        disconnect()
        
        currentEndpoint = endpoint
        
        channel = if (endpoint.startsWith("unix:")) {
            // unix socket для локальной разработки
            AndroidChannelBuilder
                .forAddress(endpoint.removePrefix("unix:"), 0)
                .usePlaintext()
                .build()
        } else {
            val parts = endpoint.split(":")
            val host = parts[0]
            val port = parts.getOrNull(1)?.toIntOrNull() ?: if (useTls) 443 else 80
            
            OkHttpChannelBuilder
                .forAddress(host, port)
                .apply {
                    if (useTls) {
                        useTransportSecurity()
                    } else {
                        usePlaintext()
                    }
                }
                .keepAliveTime(30, TimeUnit.SECONDS)
                .keepAliveTimeout(10, TimeUnit.SECONDS)
                .keepAliveWithoutCalls(true)
                .build()
        }
    }
    
    // отключиться
    suspend fun disconnect() = withContext(Dispatchers.IO) {
        channel?.shutdown()
        try {
            channel?.awaitTermination(5, TimeUnit.SECONDS)
        } catch (e: Exception) {
            channel?.shutdownNow()
        }
        channel = null
        currentEndpoint = null
    }
    
    // получить канал
    fun getChannel(): ManagedChannel? = channel
    
    // проверить подключение
    fun isConnected(): Boolean = channel != null && !channel!!.isShutdown && !channel!!.isTerminated
    
    // получить текущий endpoint
    fun getCurrentEndpoint(): String? = currentEndpoint
    
    // переподключиться
    suspend fun reconnect() = withContext(Dispatchers.IO) {
        currentEndpoint?.let { endpoint ->
            disconnect()
            connect(endpoint)
        }
    }
}
