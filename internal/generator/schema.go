// Package generator provides schema inference and fake data generation.
package generator

import (
	"regexp"
	"strings"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
)

// FieldType represents the inferred type of a field for data generation.
type FieldType int

const (
	FieldTypeUnknown FieldType = iota
	FieldTypeID
	FieldTypeEmail
	FieldTypeName
	FieldTypeFirstName
	FieldTypeLastName
	FieldTypePhone
	FieldTypeURL
	FieldTypeImage
	FieldTypeTitle
	FieldTypeDescription
	FieldTypeContent
	FieldTypeAddress
	FieldTypeCity
	FieldTypeCountry
	FieldTypeZip
	FieldTypeCompany
	FieldTypeUsername
	FieldTypePassword
	FieldTypeDate
	FieldTypePrice
	FieldTypeNumber
	FieldTypeBoolean
)

// fieldPatterns maps regex patterns to field types for intelligent inference.
var fieldPatterns = []struct {
	pattern   *regexp.Regexp
	fieldType FieldType
}{
	{regexp.MustCompile(`(?i)^id$|_id$|Id$`), FieldTypeID},
	{regexp.MustCompile(`(?i)email|e_mail|correo`), FieldTypeEmail},
	{regexp.MustCompile(`(?i)^name$|^nombre$|full_?name`), FieldTypeName},
	{regexp.MustCompile(`(?i)first_?name|primer_?nombre`), FieldTypeFirstName},
	{regexp.MustCompile(`(?i)last_?name|apellido|surname`), FieldTypeLastName},
	{regexp.MustCompile(`(?i)phone|telefono|mobile|cel`), FieldTypePhone},
	{regexp.MustCompile(`(?i)url|link|website|sitio`), FieldTypeURL},
	{regexp.MustCompile(`(?i)image|img|avatar|photo|foto`), FieldTypeImage},
	{regexp.MustCompile(`(?i)title|titulo`), FieldTypeTitle},
	{regexp.MustCompile(`(?i)desc|description|descripcion`), FieldTypeDescription},
	{regexp.MustCompile(`(?i)content|body|text|mensaje|contenido`), FieldTypeContent},
	{regexp.MustCompile(`(?i)address|direccion|street|calle`), FieldTypeAddress},
	{regexp.MustCompile(`(?i)city|ciudad`), FieldTypeCity},
	{regexp.MustCompile(`(?i)country|pais`), FieldTypeCountry},
	{regexp.MustCompile(`(?i)zip|postal|codigo_postal`), FieldTypeZip},
	{regexp.MustCompile(`(?i)company|empresa|org`), FieldTypeCompany},
	{regexp.MustCompile(`(?i)user_?name|usuario`), FieldTypeUsername},
	{regexp.MustCompile(`(?i)password|pass|clave|contrasena`), FieldTypePassword},
	{regexp.MustCompile(`(?i)date|fecha|created|updated|_at$`), FieldTypeDate},
	{regexp.MustCompile(`(?i)price|precio|cost|amount|total`), FieldTypePrice},
	{regexp.MustCompile(`(?i)count|quantity|age|num|number`), FieldTypeNumber},
	{regexp.MustCompile(`(?i)active|enabled|is_|has_|verified`), FieldTypeBoolean},
}

// InferFieldType determines the type of a field based on its name.
func InferFieldType(fieldName string) FieldType {
	for _, p := range fieldPatterns {
		if p.pattern.MatchString(fieldName) {
			return p.fieldType
		}
	}
	return FieldTypeUnknown
}

// GenerateValue creates a fake value based on the inferred field type.
func GenerateValue(fieldType FieldType) interface{} {
	switch fieldType {
	case FieldTypeID:
		return uuid.New().String()
	case FieldTypeEmail:
		return gofakeit.Email()
	case FieldTypeName:
		return gofakeit.Name()
	case FieldTypeFirstName:
		return gofakeit.FirstName()
	case FieldTypeLastName:
		return gofakeit.LastName()
	case FieldTypePhone:
		return gofakeit.Phone()
	case FieldTypeURL:
		return gofakeit.URL()
	case FieldTypeImage:
		return gofakeit.ImageURL(400, 400)
	case FieldTypeTitle:
		return gofakeit.Sentence(4)
	case FieldTypeDescription:
		return gofakeit.Sentence(10)
	case FieldTypeContent:
		return gofakeit.Paragraph(2, 3, 5, " ")
	case FieldTypeAddress:
		return gofakeit.Street()
	case FieldTypeCity:
		return gofakeit.City()
	case FieldTypeCountry:
		return gofakeit.Country()
	case FieldTypeZip:
		return gofakeit.Zip()
	case FieldTypeCompany:
		return gofakeit.Company()
	case FieldTypeUsername:
		return gofakeit.Username()
	case FieldTypePassword:
		return gofakeit.Password(true, true, true, true, false, 12)
	case FieldTypeDate:
		return gofakeit.Date().Format("2006-01-02")
	case FieldTypePrice:
		return gofakeit.Price(10, 1000)
	case FieldTypeNumber:
		return gofakeit.Number(1, 100)
	case FieldTypeBoolean:
		return gofakeit.Bool()
	default:
		return gofakeit.Word()
	}
}

// GenerateFromSample creates fake data based on a sample object's structure.
// It analyzes field names to infer types and generates appropriate fake values.
func GenerateFromSample(sample map[string]interface{}, count int) []map[string]interface{} {
	// Analyze the sample to create a schema
	schema := make(map[string]FieldType)
	for fieldName := range sample {
		schema[fieldName] = InferFieldType(fieldName)
	}

	// Generate 'count' new items
	results := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		item := make(map[string]interface{})
		for fieldName, fieldType := range schema {
			item[fieldName] = GenerateValue(fieldType)
		}
		results[i] = item
	}

	return results
}

// ExpandData takes the original data and expands each resource with generated items.
// It uses the first item of each array as a sample for field inference.
func ExpandData(data map[string]interface{}, countPerResource int) map[string]interface{} {
	expanded := make(map[string]interface{})

	for key, value := range data {
		switch v := value.(type) {
		case []interface{}:
			if len(v) > 0 {
				// Use first item as sample
				if sample, ok := v[0].(map[string]interface{}); ok {
					// Keep original items
					items := make([]interface{}, len(v))
					copy(items, v)

					// Generate and append new items
					generated := GenerateFromSample(sample, countPerResource)
					for _, gen := range generated {
						items = append(items, gen)
					}
					expanded[key] = items
				} else {
					expanded[key] = v
				}
			} else {
				expanded[key] = v
			}
		default:
			expanded[key] = v
		}
	}

	return expanded
}

// GetFieldTypeName returns a human-readable name for a field type.
func GetFieldTypeName(ft FieldType) string {
	names := map[FieldType]string{
		FieldTypeID:          "ID",
		FieldTypeEmail:       "Email",
		FieldTypeName:        "Name",
		FieldTypeFirstName:   "First Name",
		FieldTypeLastName:    "Last Name",
		FieldTypePhone:       "Phone",
		FieldTypeURL:         "URL",
		FieldTypeImage:       "Image URL",
		FieldTypeTitle:       "Title",
		FieldTypeDescription: "Description",
		FieldTypeContent:     "Content",
		FieldTypeAddress:     "Address",
		FieldTypeCity:        "City",
		FieldTypeCountry:     "Country",
		FieldTypeZip:         "Zip Code",
		FieldTypeCompany:     "Company",
		FieldTypeUsername:    "Username",
		FieldTypePassword:    "Password",
		FieldTypeDate:        "Date",
		FieldTypePrice:       "Price",
		FieldTypeNumber:      "Number",
		FieldTypeBoolean:     "Boolean",
		FieldTypeUnknown:     "Generic",
	}
	if name, ok := names[ft]; ok {
		return name
	}
	return "Unknown"
}

// AnalyzeSchema returns a map of field names to their inferred types for display.
func AnalyzeSchema(sample map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for fieldName := range sample {
		ft := InferFieldType(fieldName)
		result[fieldName] = GetFieldTypeName(ft)
	}
	return result
}

// init seeds the random number generator for gofakeit.
func init() {
	gofakeit.Seed(0) // Use current time as seed
}

// Helper to check if a string looks like a reference to another resource.
func isReferenceField(fieldName string) bool {
	lower := strings.ToLower(fieldName)
	return strings.HasSuffix(lower, "id") && len(lower) > 2
}
