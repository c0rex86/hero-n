package dev.c0rex64.heroin.core.crypto

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import java.security.KeyStore
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class SecureKeyStore @Inject constructor(
    private val context: Context
) {
    private val keyAlias = "heroin_master_key"
    private val androidKeyStore = "AndroidKeyStore"
    private val transformation = "AES/GCM/NoPadding"
    
    private val masterKey by lazy {
        MasterKey.Builder(context)
            .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
            .build()
    }
    
    private val encryptedPrefs by lazy {
        EncryptedSharedPreferences.create(
            context,
            "heroin_secure_prefs",
            masterKey,
            EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
            EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
        )
    }

    // сохранить приватный ключ
    fun storePrivateKey(alias: String, key: ByteArray) {
        encryptedPrefs.edit().putString("pk_$alias", key.toBase64()).apply()
    }

    // получить приватный ключ
    fun getPrivateKey(alias: String): ByteArray? {
        return encryptedPrefs.getString("pk_$alias", null)?.fromBase64()
    }

    // сохранить токен доступа
    fun storeAccessToken(token: String) {
        encryptedPrefs.edit().putString("access_token", token).apply()
    }

    // получить токен доступа
    fun getAccessToken(): String? {
        return encryptedPrefs.getString("access_token", null)
    }

    // сохранить refresh токен
    fun storeRefreshToken(token: String) {
        encryptedPrefs.edit().putString("refresh_token", token).apply()
    }

    // получить refresh токен
    fun getRefreshToken(): String? {
        return encryptedPrefs.getString("refresh_token", null)
    }

    // сохранить user id
    fun storeUserId(userId: String) {
        encryptedPrefs.edit().putString("user_id", userId).apply()
    }

    // получить user id
    fun getUserId(): String? {
        return encryptedPrefs.getString("user_id", null)
    }

    // очистить все данные
    fun clear() {
        encryptedPrefs.edit().clear().apply()
    }

    // генерировать aes ключ в keystore
    private fun generateKey() {
        val keyGenerator = KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES, androidKeyStore)
        val spec = KeyGenParameterSpec.Builder(
            keyAlias,
            KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
        )
            .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
            .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
            .setUserAuthenticationRequired(false)
            .build()
        keyGenerator.init(spec)
        keyGenerator.generateKey()
    }

    // получить или создать ключ
    private fun getOrCreateKey(): SecretKey {
        val keyStore = KeyStore.getInstance(androidKeyStore)
        keyStore.load(null)
        
        return if (keyStore.containsAlias(keyAlias)) {
            keyStore.getKey(keyAlias, null) as SecretKey
        } else {
            generateKey()
            keyStore.getKey(keyAlias, null) as SecretKey
        }
    }

    // шифровать данные напрямую через keystore
    fun encryptRaw(data: ByteArray): Pair<ByteArray, ByteArray> {
        val cipher = Cipher.getInstance(transformation)
        cipher.init(Cipher.ENCRYPT_MODE, getOrCreateKey())
        val iv = cipher.iv
        val encrypted = cipher.doFinal(data)
        return iv to encrypted
    }

    // расшифровать данные напрямую через keystore
    fun decryptRaw(iv: ByteArray, encrypted: ByteArray): ByteArray {
        val cipher = Cipher.getInstance(transformation)
        val spec = GCMParameterSpec(128, iv)
        cipher.init(Cipher.DECRYPT_MODE, getOrCreateKey(), spec)
        return cipher.doFinal(encrypted)
    }

    private fun ByteArray.toBase64(): String {
        return android.util.Base64.encodeToString(this, android.util.Base64.NO_WRAP)
    }

    private fun String.fromBase64(): ByteArray {
        return android.util.Base64.decode(this, android.util.Base64.NO_WRAP)
    }
}
