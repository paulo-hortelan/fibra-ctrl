package api

import (
    "fmt"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "github.com/golang-jwt/jwt/v5"
    "github.com/paulo-hortelan/fibra-ctrl/internal/queue"
    "github.com/paulo-hortelan/fibra-ctrl/internal/repository"
)

var jwtSecret = []byte("supersecret") // Use env var in production

func (s *Server) handleRegister(c *gin.Context) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
        return
    }
    hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    user := repository.User{Email: req.Email, Password: string(hash)}
    if err := s.userRepo.CreateUser(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "email already exists"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "registered"})
}

func (s *Server) handleLogin(c *gin.Context) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
        return
    }
    user, err := s.userRepo.GetUserByEmail(req.Email)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
        return
    }
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
        return
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": user.ID,
        "exp":     time.Now().Add(time.Hour * 24).Unix(),
    })
    tokenString, _ := token.SignedString(jwtSecret)
    c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func (s *Server) handleMe(c *gin.Context) {
    userID := c.GetUint("user_id")
    user, err := s.userRepo.GetUserByID(userID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"email": user.Email, "id": user.ID})
}

func (s *Server) handleLogout(c *gin.Context) {
    // For JWT, logout is handled client-side (delete token)
    c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (s *Server) authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        auth := c.GetHeader("Authorization")
        if len(auth) < 8 || auth[:7] != "Bearer " {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
            return
        }
        tokenStr := auth[7:]
        token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })
        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }
        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }
        userID, ok := claims["user_id"].(float64)
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }
        c.Set("user_id", uint(userID))
        c.Next()
    }
}