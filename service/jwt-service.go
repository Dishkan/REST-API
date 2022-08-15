package service

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"github.com/twinj/uuid"
)

//JWTService is a contract of what jwtService can do
type JWTService interface {
	GenerateToken(userID string) string
	ValidateToken(token string, tokenUuid string) (*jwt.Token, error)
	ExtractTokenMetadata(*http.Request) (*jwtCustomClaim, error)
	DeleteTokens(*jwtCustomClaim) error
}

type jwtCustomClaim struct {
	TokenUuid string
	UserID    string `json:"user_id"`
}

type jwtService struct {
	secretKey string
	issuer    string
}

//NewJWTService method is creates a new instance of JWTService
func NewJWTService() JWTService {
	return &jwtService{
		issuer:    "dishkan",
		secretKey: getSecretKey(),
	}
}

func getSecretKey() string {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey != "" {
		secretKey = "dishkan"
	}
	return secretKey
}

var client *redis.Client

func init() {
	//Initializing redis
	dsn := os.Getenv("REDIS_DSN")
	client = redis.NewClient(&redis.Options{
		Addr: dsn, //redis port
	})
	_, err := client.Ping().Result()

	if err != nil {
		panic(err)
	}
}

func (j *jwtService) GenerateToken(UserID string) string {
	TokenUuid := uuid.NewV4().String()
	ExpiresAt := time.Now().Add(time.Minute * 30).Unix()
	claims := jwt.MapClaims{}
	claims["access_uuid"] = TokenUuid
	claims["user_id"] = UserID
	claims["ExpiresAt"] = ExpiresAt
	claims["issuer"] = j.issuer
	claims["IssuedAt"] = time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	at := time.Unix(ExpiresAt, 0) //converting Unix to UTC(to Time object)
	t, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		panic(err)
	}
	client.Set(TokenUuid, t, at.Sub(time.Now())).Result()
	return t
}

func (j *jwtService) ValidateToken(token string, tokenUuid string) (*jwt.Token, error) {
	return jwt.Parse(token, func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method %v", jwtToken.Header["alg"])
		}
		redis_token, _ := client.Get(tokenUuid).Result()
		if redis_token != token {
			return nil, fmt.Errorf("Unexpected signing method %v", jwtToken.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})
}

func verifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := extractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(getSecretKey()), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

//get the token from the request body
func extractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	return bearToken
}

func extract(token *jwt.Token) (*jwtCustomClaim, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		userId, userOk := claims["user_id"].(string)
		if ok == false || userOk == false {
			return nil, errors.New("unauthorized")
		} else {
			return &jwtCustomClaim{
				TokenUuid: accessUuid,
				UserID:    userId,
			}, nil
		}
	}
	return nil, errors.New("something went wrong")
}

func (j *jwtService) ExtractTokenMetadata(r *http.Request) (*jwtCustomClaim, error) {
	token, err := verifyToken(r)
	if err != nil {
		return nil, err
	}
	acc, err := extract(token)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (j *jwtService) DeleteTokens(authD *jwtCustomClaim) error {
	//delete access token
	deletedAt, err := client.Del(authD.TokenUuid).Result()
	if err != nil {
		return err
	}
	//When the record is deleted, the return value is 1
	if deletedAt != 1 {
		return errors.New("something went wrong")
	}
	return nil
}
