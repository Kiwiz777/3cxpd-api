package main

import (
	"crypto/sha1"
	"encoding/csv"
	"encoding/hex"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kiwiz777/3cxpd-api/database"
	"gorm.io/gorm"
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
type ContactDetails struct {
	Name     string `json:"name"`
	Number   string `json:"number"`
	CFStatus string `json:"cfstatus"`
	Caller   string `json:"caller"`
	CallTime string `json:"calltime"`
	Notes    string `json:"notes"`
}

func (h *NewStorageClient) GetContacts(c *gin.Context) {
	var response []ContactDetails

	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	if pageStr == "" {
		pageStr = "1"
	}
	if limitStr == "" {
		limitStr = "30"
	}

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	offset := (page - 1) * limit

	var contacts []database.Contact

	if err := h.DB.DB.Where("cf_status = ?", "completed").Offset(offset).Limit(limit).Find(&contacts).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, contact := range contacts {
		var action database.Action
		var rec ContactDetails
		if err := h.DB.DB.Where("contact_id = ?", contact.ID).First(&action).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		rec.Name = contact.Name
		rec.Number = contact.Number
		rec.CFStatus = contact.CFStatus
		//rec.Caller = action.Caller
		rec.CallTime = action.CallTime
		rec.Notes = action.Notes

		response = append(response, rec)
	}

	log.Println(response) // Debugging log

	c.JSON(http.StatusOK, response)
}


func (h *NewStorageClient) GetContactsPending(c *gin.Context) {
	var contacts []database.Contact
	if err := h.DB.DB.Where("cf_status = ?", "pending").Find(&contacts).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, contacts)
}

func (h *NewStorageClient) NextContact(c *gin.Context) {
	var contact database.Contact
	// First check for scheduled contacts that are past their scheduled time
	now := time.Now()
	if err := h.DB.DB.Where("cf_status = ? AND is_scheduled = ? AND scheduled_time <= ?", "pending", true, now).
		Order("scheduled_time asc").
		First(&contact).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// If no scheduled contacts are ready, get first non-scheduled pending contact
			if err := h.DB.DB.Where("cf_status = ? AND is_scheduled = ?", "pending", false).
				First(&contact).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusNotFound, gin.H{"error": "No pending contacts found"})
					return
				}
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
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
        scheduledTime, err := time.Parse("2006-01-02 15:04:05", rec[3])
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format for scheduled time"})
            return
        }

        contact := database.Contact{
            ID: uuid.New(),
            Name:   rec[0],
            Number: rec[1],
            CFStatus: "pending",
            IsScheduled: rec[2] == "yes",
            ScheduledTime: scheduledTime,
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
	scheduledTime, err := time.Parse("2006-01-02 15:04:05", contact.ScheduledTime.Format("2006-01-02 15:04:05"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format for scheduled time"})
		return
	}
	contact.CFStatus = "pending"
	contact.ID = uuid.New()
	contact.ScheduledTime = scheduledTime
    if err := h.DB.DB.Create(&contact).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, contact)
}

func (h *NewStorageClient) UpdateContact(c *gin.Context) {
    number := c.GetHeader("number")
    response := c.GetHeader("response")

    if number == ""  {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Missing number or response in headers"})
        return
    }
    var req struct {
        Number string `json:"number"`
        //Caller string `json:"caller"`
        Notes  string `json:"notes"`
    }

	req.Number = number
	req.Notes = response

    var contact database.Contact
    if err := h.DB.DB.Where("number = ?", req.Number).First(&contact).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Contact not found"})
        return
    }
	if response == "false" {
		contact.CFStatus = "pending"
	} else {
		contact.CFStatus = "completed"
	}

	if err := h.DB.DB.Save(&contact).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

    action := database.Action{
		ID:        uuid.New(),
        ContactID: contact.ID,
        //Caller:    req.Caller,
        Notes:     req.Notes,
        CallTime:  time.Now().Format("2006-01-02 15:04:05"),
    }
    if err := h.DB.DB.Create(&action).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Contact updated", "action": action})
}
// delete contact
func (h *NewStorageClient) DeleteContact(c *gin.Context) {
	var contact database.Contact
	if err := h.DB.DB.Where("number = ?", c.Param("number")).First(&contact).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contact not found"})
		return
	}
	if err := h.DB.DB.Delete(&contact).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Contact deleted"})
}