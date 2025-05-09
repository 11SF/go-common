package jwt

type JWT struct{}

type IJWT interface{}

func NewJWT() IJWT {
	return JWT{}
}
