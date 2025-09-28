package crypto

import (
    "crypto/rand"
    "errors"
    "golang.org/x/crypto/chacha20poly1305"
    "golang.org/x/crypto/curve25519"
    "golang.org/x/crypto/hkdf"
    "io"
    "crypto/sha256"
)

// double ratchet state
type DoubleRatchet struct {
    // dh ratchet
    dhSend   [32]byte // наш текущий приватный ключ
    dhRecv   [32]byte // их текущий публичный ключ  
    dhPubSend [32]byte // наш текущий публичный ключ
    
    // chain keys
    rootKey  [32]byte
    sendChainKey [32]byte
    recvChainKey [32]byte
    
    // message numbers
    sendN uint32
    recvN uint32
    prevChainN uint32
    
    // skipped keys для out-of-order
    skippedKeys map[[32]byte]map[uint32][32]byte
}

// инициализация для alice (инициатор)
func InitAlice(sharedSecret, bobPub [32]byte) (*DoubleRatchet, error) {
    dr := &DoubleRatchet{
        skippedKeys: make(map[[32]byte]map[uint32][32]byte),
    }
    
    // генерим первую dh пару
    if _, err := rand.Read(dr.dhSend[:]); err != nil {
        return nil, err
    }
    curve25519.ScalarBaseMult(&dr.dhPubSend, &dr.dhSend)
    
    // первый dh
    dr.dhRecv = bobPub
    dhOut := dhExchange(dr.dhSend, dr.dhRecv)
    
    // kdf для root и send chain
    dr.rootKey, dr.sendChainKey = kdfRK(sharedSecret, dhOut)
    
    return dr, nil
}

// инициализация для bob (получатель)
func InitBob(sharedSecret [32]byte, bobPriv [32]byte) (*DoubleRatchet, error) {
    dr := &DoubleRatchet{
        dhSend: bobPriv,
        skippedKeys: make(map[[32]byte]map[uint32][32]byte),
    }
    
    curve25519.ScalarBaseMult(&dr.dhPubSend, &dr.dhSend)
    dr.rootKey = sharedSecret
    
    return dr, nil
}

// зашифровать сообщение
func (dr *DoubleRatchet) Encrypt(plaintext []byte, ad []byte) ([]byte, [32]byte, error) {
    // chain key -> message key
    var msgKey [32]byte
    msgKey, dr.sendChainKey = kdfCK(dr.sendChainKey)

    // шифруем
    header := dr.dhPubSend
    ciphertext := encrypt(msgKey, plaintext, append(ad, header[:]...))
    
    dr.sendN++
    
    return ciphertext, header, nil
}

// расшифровать сообщение
func (dr *DoubleRatchet) Decrypt(ciphertext []byte, header [32]byte, ad []byte) ([]byte, error) {
    // пробуем скипнутые ключи
    if msgKey, ok := dr.trySkipped(header, dr.recvN); ok {
        return decrypt(msgKey, ciphertext, append(ad, header[:]...))
    }

    // новый dh?
    if header != dr.dhRecv {
        if err := dr.dhRatchet(header); err != nil {
            return nil, err
        }
    }

    // chain key -> message key
    var msgKey [32]byte
    msgKey, dr.recvChainKey = kdfCK(dr.recvChainKey)

    // расшифровываем
    plaintext, err := decrypt(msgKey, ciphertext, append(ad, header[:]...))
    if err != nil {
        return nil, err
    }

    dr.recvN++

    return plaintext, nil
}

// dh ratchet step
func (dr *DoubleRatchet) dhRatchet(header [32]byte) error {
    // сохраняем скипнутые
    dr.skipMessageKeys(dr.dhRecv, dr.prevChainN)

    // новая receive chain
    dr.prevChainN = dr.sendN
    dr.sendN = 0
    dr.recvN = 0
    dr.dhRecv = header

    // receive dh
    dhOut := dhExchange(dr.dhSend, dr.dhRecv)
    dr.rootKey, dr.recvChainKey = kdfRK(dr.rootKey, dhOut)

    // новая send dh пара
    if _, err := rand.Read(dr.dhSend[:]); err != nil {
        return err
    }
    curve25519.ScalarBaseMult(&dr.dhPubSend, &dr.dhSend)

    // send dh
    dhOut = dhExchange(dr.dhSend, dr.dhRecv)
    dr.rootKey, dr.sendChainKey = kdfRK(dr.rootKey, dhOut)

    return nil
}

// сохранить скипнутые ключи
func (dr *DoubleRatchet) skipMessageKeys(pub [32]byte, until uint32) {
    if dr.recvN+100 < until {
        return // слишком много, защита от dos
    }
    
    if _, ok := dr.skippedKeys[pub]; !ok {
        dr.skippedKeys[pub] = make(map[uint32][32]byte)
    }
    
    for dr.recvN < until {
        msgKey, chainKey := kdfCK(dr.recvChainKey)
        dr.skippedKeys[pub][dr.recvN] = msgKey
        dr.recvChainKey = chainKey
        dr.recvN++
    }
}

// проверить скипнутые
func (dr *DoubleRatchet) trySkipped(header [32]byte, nr uint32) ([32]byte, bool) {
    if keys, ok := dr.skippedKeys[header]; ok {
        if msgKey, ok := keys[nr]; ok {
            delete(keys, nr)
            if len(keys) == 0 {
                delete(dr.skippedKeys, header)
            }
            return msgKey, true
        }
    }
    return [32]byte{}, false
}

// dh exchange
func dhExchange(priv, pub [32]byte) [32]byte {
    var shared [32]byte
    curve25519.ScalarMult(&shared, &priv, &pub)
    return shared
}

// kdf для root key
func kdfRK(rk, dhOut [32]byte) ([32]byte, [32]byte) {
    var newRK, chainKey [32]byte
    
    kdf := hkdf.New(sha256.New, dhOut[:], rk[:], []byte("heroin_ratchet"))
    io.ReadFull(kdf, newRK[:])
    io.ReadFull(kdf, chainKey[:])
    
    return newRK, chainKey
}

// kdf для chain key
func kdfCK(ck [32]byte) ([32]byte, [32]byte) {
    var msgKey, newCK [32]byte
    
    // простой hmac-based kdf
    h1 := sha256.Sum256(append(ck[:], 0x01))
    h2 := sha256.Sum256(append(ck[:], 0x02))
    
    copy(msgKey[:], h1[:])
    copy(newCK[:], h2[:])
    
    return msgKey, newCK
}

// шифрование xchacha20-poly1305
func encrypt(key [32]byte, plaintext, ad []byte) []byte {
    aead, _ := chacha20poly1305.NewX(key[:])
    
    nonce := make([]byte, aead.NonceSize())
    rand.Read(nonce)
    
    ciphertext := aead.Seal(nonce, nonce, plaintext, ad)
    return ciphertext
}

// расшифровка xchacha20-poly1305
func decrypt(key [32]byte, ciphertext, ad []byte) ([]byte, error) {
    aead, _ := chacha20poly1305.NewX(key[:])
    
    if len(ciphertext) < aead.NonceSize() {
        return nil, errors.New("ciphertext too short")
    }
    
    nonce := ciphertext[:aead.NonceSize()]
    ciphertext = ciphertext[aead.NonceSize():]
    
    return aead.Open(nil, nonce, ciphertext, ad)
}
