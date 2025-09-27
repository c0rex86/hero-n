package dev.c0rex64.heroin.core.crypto

import com.goterl.lazysodium.LazySodiumAndroid
import com.goterl.lazysodium.SodiumAndroid
import com.goterl.lazysodium.interfaces.AEAD
import com.goterl.lazysodium.interfaces.Sign
import com.goterl.lazysodium.interfaces.KeyExchange
import com.goterl.lazysodium.interfaces.PasswordHash
import com.goterl.lazysodium.utils.Key
import com.goterl.lazysodium.utils.LibraryLoader
import com.goterl.lazysodium.utils.SodiumLibrary

class Crypto {
    private val sodium: LazySodiumAndroid
    private val aead: AEAD
    private val sign: Sign
    private val kx: KeyExchange
    private val pwh: PasswordHash

    init {
        LibraryLoader.load(SodiumLibrary.LIB_NAME)
        val sa = SodiumAndroid()
        sodium = LazySodiumAndroid(sa)
        aead = sodium
        sign = sodium
        kx = sodium
        pwh = sodium
    }

    fun xchachaSeal(key: ByteArray, nonce: ByteArray, plaintext: ByteArray, aad: ByteArray? = null): ByteArray {
        require(key.size == AEAD.XCHACHA20POLY1305_IETF_KEYBYTES)
        require(nonce.size == AEAD.XCHACHA20POLY1305_IETF_NPUBBYTES)
        val out = ByteArray(plaintext.size + AEAD.XCHACHA20POLY1305_IETF_ABYTES)
        val ok = aead.cryptoAeadXChaCha20Poly1305IetfEncrypt(out, null, plaintext, aad, null, nonce, key)
        check(ok) { "encryption failed" }
        return out
    }

    fun xchachaOpen(key: ByteArray, nonce: ByteArray, ciphertext: ByteArray, aad: ByteArray? = null): ByteArray {
        require(key.size == AEAD.XCHACHA20POLY1305_IETF_KEYBYTES)
        require(nonce.size == AEAD.XCHACHA20POLY1305_IETF_NPUBBYTES)
        val out = ByteArray(ciphertext.size - AEAD.XCHACHA20POLY1305_IETF_ABYTES)
        val ok = aead.cryptoAeadXChaCha20Poly1305IetfDecrypt(out, null, null, ciphertext, aad, nonce, key)
        check(ok) { "decryption failed" }
        return out
    }

    fun ed25519Keypair(): Pair<ByteArray, ByteArray> {
        val pk = ByteArray(Sign.PUBLICKEYBYTES)
        val sk = ByteArray(Sign.SECRETKEYBYTES)
        val ok = sign.cryptoSignKeypair(pk, sk)
        check(ok) { "keypair failed" }
        return pk to sk
    }

    fun ed25519Sign(sk: ByteArray, message: ByteArray): ByteArray {
        val sig = ByteArray(Sign.BYTES)
        val ok = sign.cryptoSignDetached(sig, null, message, message.size.toLong(), sk)
        check(ok) { "sign failed" }
        return sig
    }

    fun ed25519Verify(pk: ByteArray, message: ByteArray, sig: ByteArray): Boolean {
        if (pk.size != Sign.PUBLICKEYBYTES || sig.size != Sign.BYTES) return false
        return sign.cryptoSignVerifyDetached(sig, message, message.size.toLong(), pk)
    }

    fun x25519Keypair(): Pair<ByteArray, ByteArray> {
        val pk = ByteArray(KeyExchange.PUBLICKEYBYTES)
        val sk = ByteArray(KeyExchange.SECRETKEYBYTES)
        val ok = kx.cryptoKxKeypair(pk, sk)
        check(ok) { "kx keypair failed" }
        return pk to sk
    }

    fun argon2id(password: ByteArray, salt: ByteArray, ops: Long = PasswordHash.OPSLIMIT_MODERATE, mem: Long = PasswordHash.MEMLIMIT_MODERATE, outLen: Int = 32): ByteArray {
        val out = ByteArray(outLen)
        val ok = pwh.cryptoPwHash(out, outLen.toLong(), password, password.size.toLong(), salt, ops, mem, PasswordHash.Alg.PWHASH_ALG_ARGON2ID13)
        check(ok) { "argon2id failed" }
        return out
    }

    fun random(bytes: Int): ByteArray {
        val k = Key.fromBytes(ByteArray(bytes))
        sodium.cryptoSecureRandom(k.asBytes)
        return k.asBytes
    }
}
