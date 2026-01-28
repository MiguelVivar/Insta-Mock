// Package generator provides schema-based data generation.
package generator

import (
	"strings"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
)

// BuildFromSchema generates a realistic JSON map based on a schema definition.
// Schema is a map where keys are field paths (dot notation for nested) and values are type names.
// Example: {"name": "name", "email": "email", "address.city": "city", "address.zip": "zip"}
func BuildFromSchema(schema map[string]string) map[string]interface{} {
	result := make(map[string]interface{})

	for path, fieldType := range schema {
		value := GenerateByType(fieldType)
		setNestedValue(result, path, value)
	}

	return result
}

// BuildManyFromSchema generates N items from a schema.
func BuildManyFromSchema(schema map[string]string, count int) []map[string]interface{} {
	results := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		results[i] = BuildFromSchema(schema)
	}
	return results
}

// GenerateByType returns a fake value based on the type string.
func GenerateByType(fieldType string) interface{} {
	switch strings.ToLower(fieldType) {
	// Identity
	case "uuid", "id":
		return uuid.New().String()
	case "name", "fullname", "full_name":
		return gofakeit.Name()
	case "firstname", "first_name":
		return gofakeit.FirstName()
	case "lastname", "last_name":
		return gofakeit.LastName()
	case "username", "user_name":
		return gofakeit.Username()

	// Contact
	case "email", "mail":
		return gofakeit.Email()
	case "phone", "telephone", "mobile":
		return gofakeit.Phone()
	case "url", "website", "link":
		return gofakeit.URL()

	// Network
	case "ipv4", "ip":
		return gofakeit.IPv4Address()
	case "ipv6":
		return gofakeit.IPv6Address()
	case "mac", "macaddress":
		return gofakeit.MacAddress()
	case "useragent", "user_agent":
		return gofakeit.UserAgent()

	// Business
	case "company", "empresa":
		return gofakeit.Company()
	case "jobtitle", "job_title", "job":
		return gofakeit.JobTitle()
	case "currency", "currency_code":
		return gofakeit.CurrencyShort()
	case "price", "amount", "money":
		return gofakeit.Price(10, 1000)
	case "creditcard", "credit_card", "cc":
		return gofakeit.CreditCardNumber(nil)

	// Location
	case "address", "street":
		return gofakeit.Street()
	case "city", "ciudad":
		return gofakeit.City()
	case "state", "estado":
		return gofakeit.State()
	case "country", "pais":
		return gofakeit.Country()
	case "zip", "zipcode", "postal":
		return gofakeit.Zip()
	case "latitude", "lat":
		return gofakeit.Latitude()
	case "longitude", "lng", "lon":
		return gofakeit.Longitude()

	// Text
	case "word":
		return gofakeit.Word()
	case "sentence":
		return gofakeit.Sentence(8)
	case "paragraph":
		return gofakeit.Paragraph(2, 3, 5, " ")
	case "title":
		return gofakeit.Sentence(4)
	case "description", "desc":
		return gofakeit.Sentence(10)
	case "lorem":
		return gofakeit.LoremIpsumSentence(10)

	// Numbers
	case "random_int", "int", "integer", "number":
		return gofakeit.Number(1, 1000)
	case "float", "decimal":
		return gofakeit.Float64Range(1, 1000)
	case "age":
		return gofakeit.Number(18, 80)
	case "year":
		return gofakeit.Year()

	// Boolean
	case "bool", "boolean":
		return gofakeit.Bool()

	// Date/Time
	case "date":
		return gofakeit.Date().Format("2006-01-02")
	case "datetime", "timestamp":
		return gofakeit.Date().Format("2006-01-02T15:04:05Z")
	case "time":
		return gofakeit.Date().Format("15:04:05")

	// Media
	case "image", "avatar", "photo":
		return gofakeit.ImageURL(400, 400)
	case "color", "hex_color":
		return gofakeit.HexColor()
	case "rgb":
		return gofakeit.RGBColor()

	// Tech
	case "password":
		return gofakeit.Password(true, true, true, true, false, 12)
	case "hash", "md5":
		return gofakeit.UUID() // Simplified
	case "domain":
		return gofakeit.DomainName()

	// Default: random word
	default:
		return gofakeit.Word()
	}
}

// setNestedValue sets a value in a nested map using dot notation path.
// Example: setNestedValue(m, "address.city", "NYC") creates m["address"]["city"] = "NYC"
func setNestedValue(m map[string]interface{}, path string, value interface{}) {
	parts := strings.Split(path, ".")

	// Navigate/create nested structure
	current := m
	for i := 0; i < len(parts)-1; i++ {
		key := parts[i]
		if _, exists := current[key]; !exists {
			current[key] = make(map[string]interface{})
		}
		if nested, ok := current[key].(map[string]interface{}); ok {
			current = nested
		} else {
			// Key exists but is not a map, overwrite with map
			newMap := make(map[string]interface{})
			current[key] = newMap
			current = newMap
		}
	}

	// Set the final value
	current[parts[len(parts)-1]] = value
}

// GetNestedValue retrieves a value from a nested map using dot notation.
func GetNestedValue(m map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")

	current := m
	for i := 0; i < len(parts)-1; i++ {
		if nested, ok := current[parts[i]].(map[string]interface{}); ok {
			current = nested
		} else {
			return nil, false
		}
	}

	value, exists := current[parts[len(parts)-1]]
	return value, exists
}

// MergeSchema combines two schemas, with the second schema taking precedence.
func MergeSchema(base, override map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		result[k] = v
	}
	return result
}

// CommonSchemas provides predefined schemas for common entities.
var CommonSchemas = map[string]map[string]string{
	"user": {
		"id":         "uuid",
		"name":       "name",
		"email":      "email",
		"phone":      "phone",
		"avatar":     "image",
		"created_at": "datetime",
	},
	"product": {
		"id":          "uuid",
		"name":        "word",
		"description": "description",
		"price":       "price",
		"image":       "image",
		"category":    "word",
	},
	"post": {
		"id":         "uuid",
		"title":      "title",
		"content":    "paragraph",
		"author":     "name",
		"created_at": "datetime",
	},
	"company": {
		"id":              "uuid",
		"name":            "company",
		"email":           "email",
		"phone":           "phone",
		"address.street":  "address",
		"address.city":    "city",
		"address.country": "country",
		"address.zip":     "zip",
	},
}

// GenerateFromCommonSchema generates items using a predefined common schema.
func GenerateFromCommonSchema(schemaName string, count int) []map[string]interface{} {
	schema, exists := CommonSchemas[schemaName]
	if !exists {
		return nil
	}
	return BuildManyFromSchema(schema, count)
}
