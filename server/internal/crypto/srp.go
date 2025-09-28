package crypto

import (
    "crypto/rand"
    "crypto/sha256"
    "errors"
    "math/big"
)

var (
    N = fromHex("EEAF0AB9ADB38DD69C33F80AFA8FC5E86072618775FF3C0B9EA2314C9C256576D674DF7496EA81D3383B4813D692C6E0E0D5D8E250B98BE48E495C1D6089DAD15DC7D7B46154D6B6CE8EF4AD69B15D4982559B297BCF1885C529F566660E57EC68EDBC3C05726CC02FD4CBF4976EAA9AFD5138FE8376435B9FC61D2FC0EB06E3")
    g = big.NewInt(2)
    k = fromHex("5b9e8ef059c6b32ea59fc1d322d37f04aa30bae5aa9003b8321e21ddb04e300")
)

type SRP struct {
    N *big.Int
    g *big.Int
    k *big.Int
}

func NewSRP() *SRP {
    return &SRP{
        N: N,
        g: g,
        k: k,
    }
}

type ClientSession struct {
    srp      *SRP
    username string
    a        *big.Int
    A        *big.Int
    K        []byte
    M1       []byte
}

type ServerSession struct {
    srp      *SRP
    username string
    salt     []byte
    verifier *big.Int
    b        *big.Int
    B        *big.Int
    K        []byte
    M1       []byte
}

func (s *SRP) NewClientSession(username string, password string) (*ClientSession, error) {
    a, err := rand.Int(rand.Reader, s.N)
    if err != nil {
        return nil, err
    }
    
    A := new(big.Int).Exp(s.g, a, s.N)
    
    return &ClientSession{
        srp:      s,
        username: username,
        a:        a,
        A:        A,
    }, nil
}

func (c *ClientSession) ProcessChallenge(salt []byte, B *big.Int) ([]byte, error) {
    if B.Mod(B, c.srp.N).Cmp(big.NewInt(0)) == 0 {
        return nil, errors.New("invalid B")
    }
    
    u := hashBigInts(c.A, B)
    
    x := hashPassword(c.username, salt, c.username)
    
    kgx := new(big.Int).Mul(c.srp.k, new(big.Int).Exp(c.srp.g, x, c.srp.N))
    diff := new(big.Int).Sub(B, kgx)
    diff.Mod(diff, c.srp.N)
    
    ux := new(big.Int).Mul(u, x)
    aux := new(big.Int).Add(c.a, ux)
    
    S := new(big.Int).Exp(diff, aux, c.srp.N)
    
    c.K = hash(S.Bytes())
    
    c.M1 = hashM1(c.username, salt, c.A, B, c.K)
    
    return c.M1, nil
}

func (c *ClientSession) VerifyServerProof(M2 []byte) bool {
    expected := hashM2(c.A, c.M1, c.K)
    return bytesEqual(M2, expected)
}

func (s *SRP) NewServerSession(username string, salt []byte, verifier *big.Int) (*ServerSession, error) {
    b, err := rand.Int(rand.Reader, s.N)
    if err != nil {
        return nil, err
    }
    
    gb := new(big.Int).Exp(s.g, b, s.N)
    kv := new(big.Int).Mul(s.k, verifier)
    B := new(big.Int).Add(kv, gb)
    B.Mod(B, s.N)
    
    return &ServerSession{
        srp:      s,
        username: username,
        salt:     salt,
        verifier: verifier,
        b:        b,
        B:        B,
    }, nil
}

func (s *ServerSession) VerifyClientProof(A *big.Int, M1 []byte) ([]byte, error) {
    if A.Mod(A, s.srp.N).Cmp(big.NewInt(0)) == 0 {
        return nil, errors.New("invalid A")
    }
    
    u := hashBigInts(A, s.B)
    
    vu := new(big.Int).Exp(s.verifier, u, s.srp.N)
    Avu := new(big.Int).Mul(A, vu)
    
    S := new(big.Int).Exp(Avu, s.b, s.srp.N)
    
    s.K = hash(S.Bytes())
    
    expectedM1 := hashM1(s.username, s.salt, A, s.B, s.K)
    if !bytesEqual(M1, expectedM1) {
        return nil, errors.New("invalid proof")
    }
    
    s.M1 = M1
    
    M2 := hashM2(A, M1, s.K)
    return M2, nil
}

func (s *SRP) CreateVerifier(username, password string, salt []byte) *big.Int {
    x := hashPassword(username, salt, password)
    return new(big.Int).Exp(s.g, x, s.N)
}

func hashPassword(username string, salt []byte, password string) *big.Int {
    h := sha256.New()
    h.Write([]byte(username))
    h.Write([]byte(":"))
    h.Write([]byte(password))
    inner := h.Sum(nil)
    
    h.Reset()
    h.Write(salt)
    h.Write(inner)
    
    return new(big.Int).SetBytes(h.Sum(nil))
}

func hashBigInts(a, b *big.Int) *big.Int {
    h := sha256.New()
    h.Write(a.Bytes())
    h.Write(b.Bytes())
    return new(big.Int).SetBytes(h.Sum(nil))
}

func hashM1(username string, salt []byte, A, B *big.Int, K []byte) []byte {
    h := sha256.New()
    h.Write([]byte(username))
    h.Write(salt)
    h.Write(A.Bytes())
    h.Write(B.Bytes())
    h.Write(K)
    return h.Sum(nil)
}

func hashM2(A *big.Int, M1, K []byte) []byte {
    h := sha256.New()
    h.Write(A.Bytes())
    h.Write(M1)
    h.Write(K)
    return h.Sum(nil)
}

func hash(data []byte) []byte {
    h := sha256.Sum256(data)
    return h[:]
}

func bytesEqual(a, b []byte) bool {
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}

func fromHex(s string) *big.Int {
    n, _ := new(big.Int).SetString(s, 16)
    return n
}
