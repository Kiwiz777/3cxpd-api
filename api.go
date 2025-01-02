package main

import (
	"crypto/sha1"
	"encoding/csv"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kiwiz777/3cxpd-api/database"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type CheckTokenRequest struct {
	Token string `json:"token"`
}

func (h *NewStorageClient) Login(c *gin.Context) {
	var request LoginRequest
	var user database.User
	var token database.Token
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hash := sha1.Sum([]byte(request.Password))
	if err := h.DB.DB.Where("username = ? AND password = ?", request.Username, hex.EncodeToString(hash[:])).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password"})
		return
	}
	if err := h.DB.DB.Where("user_id = ?", user.ID).First(&token).Error; err != nil {
		token.Token = database.GenerateToken(32)
		token.UserID = user.ID
		token.ID = uuid.New()
		h.DB.DB.Create(&token)
	}
	c.JSON(http.StatusOK, gin.H{"token": token.Token})
}
func (h *NewStorageClient) CheckAuth(c *gin.Context) {
    authorization := c.GetHeader("Authorization")

    // Check for valid token
    var token database.Token
    if err := h.DB.DB.Where("token = ?", authorization).First(&token).Error; err == nil {
        c.Next()
        return
    }

    // Check for valid system key
    var sysKey database.SystemKey
    if err := h.DB.DB.Where("key = ?", authorization).First(&sysKey).Error; err == nil {
        c.Next()
        return
    }

    c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token or key"})
    c.Abort()
}
func (h *NewStorageClient) CheckToken(c *gin.Context) {
	var token database.Token
	if err := h.DB.DB.Where("token = ?", c.GetHeader("Authorization")).First(&token).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *NewStorageClient) GetContacts(c *gin.Context) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	if pageStr == "" {
		pageStr = "1"
	}
	if limitStr == "" {
		limitStr = "10"
	}

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	offset := (page - 1) * limit

	var contacts []database.Contact
	if err := h.DB.DB.Offset(offset).Limit(limit).Find(&contacts).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contacts)
}

func (h *NewStorageClient) GetContactsPending(c *gin.Context) {
	var contacts []database.Contact
	if err := h.DB.DB.Where("cfstatus = ?", "pending").Find(&contacts).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contacts)
}

func (h *NewStorageClient) NextContact(c *gin.Context) {
	var contact database.Contact
	if err := h.DB.DB.Where("cfstatus = ?", "pending").First(&contact).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	contact.CFStatus = "inprogress"
	if err := h.DB.DB.Save(&contact).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contact)
}

func (h *NewStorageClient) BulkAddContacts(c *gin.Context) {
    file, _, err := c.Request.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    defer file.Close()

    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    for _, rec := range records {
        contact := database.Contact{
            Name:   rec[0],
            Number: rec[1],
            // other fields...
        }
        if err := h.DB.DB.Create(&contact).Error; err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
    }
    c.JSON(http.StatusOK, gin.H{"message": "Contacts added"})
}

func (h *NewStorageClient) AddContact(c *gin.Context) {
    var contact database.Contact
    if err := c.BindJSON(&contact); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if err := h.DB.DB.Create(&contact).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, contact)
}

func (h *NewStorageClient) UpdateContact(c *gin.Context) {
    var req struct {
        Number string `json:"number"`
        Caller string `json:"caller"`
        Notes  string `json:"notes"`
    }
    if err := c.BindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var contact database.Contact
    if err := h.DB.DB.Where("number = ?", req.Number).First(&contact).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Contact not found"})
        return
    }
	contact.CFStatus = "completed"
	if err := h.DB.DB.Save(&contact).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

    action := database.Action{
        ContactID: contact.ID,
        Caller:    req.Caller,
        Notes:     req.Notes,
        CallTime:  time.Now().Format("2006-01-02 15:04:05"),
    }
    if err := h.DB.DB.Create(&action).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Contact updated", "action": action})
}
