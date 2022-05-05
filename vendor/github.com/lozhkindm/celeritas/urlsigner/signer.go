package urlsigner

import (
	"fmt"
	"strings"
	"time"

	goalone "github.com/bwmarrin/go-alone"
)

type Signer struct {
	Secret []byte
}

func (s *Signer) GenerateTokenFromString(url string) string {
	var sign string
	if strings.Contains(url, "?") {
		sign = fmt.Sprintf("%s&hash=", url)
	} else {
		sign = fmt.Sprintf("%s?hash=", url)
	}
	crypt := goalone.New(s.Secret, goalone.Timestamp)
	return string(crypt.Sign([]byte(sign)))
}

func (s *Signer) VerifyToken(token string) bool {
	crypt := goalone.New(s.Secret, goalone.Timestamp)
	_, err := crypt.Unsign([]byte(token))
	return err == nil
}

func (s *Signer) Expired(token string, ttl time.Duration) bool {
	crypt := goalone.New(s.Secret, goalone.Timestamp)
	sign := crypt.Parse([]byte(token))
	return time.Since(sign.Timestamp) > ttl
}
