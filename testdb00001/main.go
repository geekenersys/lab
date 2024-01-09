package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/howeyc/crc16"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v4"
)

// QRRequest structure
type QRRequest struct {
	ID            string `gorm:"primaryKey"`
	TxID          string
	Type          string
	RecipientID   string
	RecipientType string
	MerchantName  string
	Reference1    string
	Reference2    string
	Amount        float64
	Onetime       bool
	Remark        string
	CreatedAt     int64
	QRCode        string
	Expire        int64
}

// Config structure
type Config struct {
	DBHost     string
	DBName     string
	DBPort     string
	DBUser     string
	DBPassword string
	ServerPort string
}

// LoadConfig function
func LoadConfig() (*Config, error) {
	config := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBName:     os.Getenv("DB_NAME"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		ServerPort: os.Getenv("SERVER_PORT"),
	}

	return config, nil
}

var jwtSecretKey = []byte("3ab92c27e5d24fe682e73b3a9d9c2a62")

func login(c *fiber.Ctx) error {

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = "enersys"
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": t})
}

// setupDatabaseConnection function
func setupDatabaseConnection(cfg *Config) (*gorm.DB, error) {
	dsn := "host=" + cfg.DBHost + " user=" + cfg.DBUser + " password=" + cfg.DBPassword + " dbname=" + cfg.DBName + " port=" + cfg.DBPort + " sslmode=disable TimeZone=Asia/Bangkok"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&QRRequest{})

	return db, nil
}

// CRUD functions
func CreateQRRequest(db *gorm.DB, qr *QRRequest) error {
	return db.Create(qr).Error
}

func GetQRRequest(db *gorm.DB, id string) (*QRRequest, error) {
	var qr QRRequest
	result := db.First(&qr, "id = ?", id)
	return &qr, result.Error
}

// ... additional CRUD functions for Update and Delete ...

func createQRRequestHandler(db *gorm.DB) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		qr := new(QRRequest)
		if err := c.BodyParser(qr); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		// Validate the amount
		if err := validateAmount(qr.Amount); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		// Proceed with creating the QR request
		if err := CreateQRRequest(db, qr); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(qr)
	}
}

// Route handlers
// func createQRRequestHandler(db *gorm.DB) func(*fiber.Ctx) error {
// 	return func(c *fiber.Ctx) error {
// 		qr := new(QRRequest)
// 		if err := c.BodyParser(qr); err != nil {
// 			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
// 		}

// 		if err := CreateQRRequest(db, qr); err != nil {
// 			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
// 		}

// 		return c.JSON(qr)
// 	}
// }

func getQRRequestHandler(db *gorm.DB) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		qr, err := GetQRRequest(db, id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "QRRequest not found"})
		}

		return c.JSON(qr)
	}
}

// ... additional handlers for Update and Delete ...
func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := setupDatabaseConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer closeDatabaseConnection(db)

	app := fiber.New()

	// Configure global middleware here (if any)

	handler := NewHandler(db)
	setupRoutes(app, db, handler)

	log.Printf("Starting server on port %s", cfg.ServerPort)
	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func MockupDB(db *gorm.DB) error {
	return db.AutoMigrate(&QRRequest{})
}

func InjectTestData(db *gorm.DB) error {
	testData := []QRRequest{
		// Populate with test data
		{
			TxID: "InjectTestData-TestTx1", Type: "InjectTestData-TestType1", RecipientID: "InjectTestData-TestRec1", /* other fields */
		},
		// Add more test data as needed
	}

	for _, data := range testData {
		if err := db.Create(&data).Error; err != nil {
			return err
		}
	}
	return nil
}

func closeDatabaseConnection(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Error closing database: %v", err)
	} else {
		sqlDB.Close()
	}
}

// func setupRoutes(app *fiber.App, myHandler *Handler) {
// 	app.Post("/generateqr", jwtware.New(jwtware.Config{SigningKey: jwtSecretKey}), myHandler.generateQR)
// 	// ... other routes ...
// }

// func setupRoutes(app *fiber.App, myHandler *Handler) {
// 	// Existing routes
// 	app.Post("/generateqr", jwtware.New(jwtware.Config{SigningKey: jwtSecretKey}), myHandler.generateQR)
// 	// ... other routes ...

// 	// Add the login route
// 	app.Post("/login", login)

// 	// You can also add other routes here, like health check or a default route
// 	app.Get("/health", func(c *fiber.Ctx) error {
// 		return c.JSON(fiber.Map{"status": "ok"})
// 	})

// 	app.Get("/", func(c *fiber.Ctx) error {
// 		return c.SendString("QR generator service up!")
// 	})

// 	// ... additional routes for Update, Delete, etc.
// }

func setupRoutes(app *fiber.App, db *gorm.DB, handler *Handler) {
	// Existing routes...
	app.Post("/generateqr", jwtware.New(jwtware.Config{SigningKey: jwtSecretKey}), handler.generateQR)
	// ...

	// Route to set up the database schema
	app.Post("/mockupdb", func(c *fiber.Ctx) error {
		if err := MockupDB(db); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendString("Database schema setup completed.")
	})

	// Route to inject test data into the database
	app.Post("/injecttestdata", func(c *fiber.Ctx) error {
		if err := InjectTestData(db); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendString("Test data injection completed.")
	})

	// Other routes...
	app.Post("/login", login)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("QR generator service up!")
	})
	// ...
}

func (h *Handler) generateQR(c *fiber.Ctx) error {
	data := new(RequestData)
	if err := c.BodyParser(data); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	// Generate QR code data
	qrCodeData := GenerateBillPaymentQRCode(uuid.New().String(), data.BillerId, data.MerchantName, data.Reference1, data.Reference2, data.Amount, data.Onetime)

	// Create and populate a new QRRequest instance
	qrRequest := &QRRequest{
		// Assign appropriate values to each field
		ID:            uuid.New().String(),
		TxID:          data.TxId,
		Type:          "promptpay",
		RecipientID:   data.RecipientId,
		RecipientType: data.RecipientType,
		MerchantName:  data.MerchantName,
		Reference1:    data.Reference1,
		Reference2:    data.Reference2,
		Amount:        data.Amount,
		Onetime:       data.Onetime,
		Remark:        data.Remark,
		CreatedAt:     time.Now().Unix(),
		QRCode:        qrCodeData,
		Expire:        data.Expire,
		// ... other necessary fields ...
	}

	// Save the QRRequest to the database
	if err := CreateQRRequest(h.db, qrRequest); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(qrRequest)
}

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) ExecuteJob(args ...interface{}) (interface{}, error) {
	// Placeholder return statement, adjust according to your logic
	return nil, nil
}

type RequestData struct {
	BillerId      string  `json:"billerId"`
	MerchantName  string  `json:"merchantName"`
	Reference1    string  `json:"reference1"`
	Reference2    string  `json:"reference2"`
	Amount        float64 `json:"amount"`
	Onetime       bool    `json:"onetime"`
	TxId          string  `json:"txId"`          // Transaction ID
	RecipientId   string  `json:"recipientId"`   // Recipient ID
	RecipientType string  `json:"recipientType"` // Recipient Type
	Remark        string  `json:"remark"`        // Additional Remark
	Expire        int64   `json:"expire"`        // Expiration time
}

// type RequestData struct {
// 	BillerId     string  `json:"billerId"`
// 	MerchantName string  `json:"merchantName"`
// 	Reference1   string  `json:"reference1"`
// 	Reference2   string  `json:"reference2"`
// 	Amount       float64 `json:"amount"`
// 	Onetime      bool    `json:"onetime"`
// }

func formatAmountForQR(amount float64) string {
	return fmt.Sprintf("%.2f", amount) // Format as a string with two decimal places
}

func formatQRField(prefix, value string) string {
	if len(value) != 0 {
		if len(value) < 10 {
			value = prefix + "0" + strconv.Itoa(len(value)) + value
		} else {
			value = prefix + strconv.Itoa(len(value)) + value
		}
	} else {
		value = ""
	}
	return value
}

func formatReceiverIDForQR(value string) string {
	checkAmount := strings.Split(value, ".")

	if len(checkAmount) > 1 {
		if checkAmount[1] == "" || len(checkAmount[1]) == 0 {
			checkAmount[1] = "00"
		} else if len(checkAmount[1]) == 1 {
			checkAmount[1] += "0"
		} else if len(checkAmount[1]) > 2 {
			checkAmount[1] = checkAmount[1][:2]
		}

		value = checkAmount[0] + "." + checkAmount[1]
	} else if len(checkAmount) == 1 {
		value = value + "." + "00"
	}

	value = formatQRField("54", value)

	return value
}

func removeFirstRune(value string) string {
	if len(value) == 10 && value[0] == '0' { // Phone Number
		value = "0066" + trimFirstRune(value)
		value = formatQRField("01", value)
	} else if len(value) == 13 { // National ID or Tax ID
		value = formatQRField("02", value)
	} else if len(value) == 15 { // E-Wallet ID
		value = formatQRField("03", value)
	} else { // Bank Account
		value = formatQRField("04", value)
	}

	return value
}

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}

func GenerateBillPaymentQRCode(qrID, billerId, merchantName, reference1, reference2 string, amount float64, onetime bool) string {
	amountStr := formatReceiverIDForQR(fmt.Sprintf("%.2f", amount)) // Convert amount to string in QR format

	reference1 = strings.ToUpper(reference1)
	reference2 = strings.ToUpper(reference2)

	PFI := formatQRField("00", "01")

	var pim_val string
	if onetime {
		pim_val = "12"
	} else {
		pim_val = "11"
	}
	PIM := formatQRField("01", pim_val)

	/** Merchant Identifier */
	AID := formatQRField("00", "A000000677010112")
	billerId = formatQRField("01", billerId)
	reference1 = formatQRField("02", reference1)
	reference2 = formatQRField("03", reference2)

	merchantSum := AID + billerId + reference1 + reference2
	merchantIdentifier := formatQRField("30", merchantSum)
	/* */

	Currency := formatQRField("53", "764")

	CountryCode := formatQRField("58", "TH")

	merchantName = formatQRField("59", merchantName)

	TerminalID := formatQRField("07", "")

	crc := "6304"

	data := PFI + PIM + merchantIdentifier + Currency + amountStr + CountryCode + merchantName + TerminalID + crc
	dataBuffer := []byte(data)
	crcResult := crc16.ChecksumCCITTFalse(dataBuffer)

	data += strings.ToUpper(strconv.FormatInt(int64(crcResult), 16))

	return data
}

func validateAmount(amount float64) error {
	// Check if the amount is within the desired range
	if amount <= 0 || amount > 2000000000000000 {
		return fmt.Errorf("amount must be greater than 0 and less than or equal to 2,000,000,000,000,000")
	}

	// Check if the amount has two decimal places
	if amount != math.Floor(amount*100)/100 {
		return fmt.Errorf("amount must have two decimal places")
	}

	return nil
}
