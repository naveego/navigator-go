package pipeline

import (
	"errors"
	"hash/crc32"
	"sort"
	"strings"
	"time"
	"unicode"
)

var castagnoliTable = crc32.MakeTable(crc32.Castagnoli) // see http://golang.org/pkg/hash/crc32/#pkg-constants

// Shape is used to maintain type information about the data contained in the dataPoint.  Shape information
// may be provided from the producer, but it is not required.  The pipeline will generate type information
// automatically based on the data itself.
type Shape struct {
	KeyNames     []string `json:"keyNames,omitempty"`     // An array of key property names
	KeyNamesHash uint32   `json:"keyNamesHash,omitempty"` // A hash used to determine if keys have changed
	Properties   []string `json:"properties,omitempty"`   // An array of properties including type, the form of [name]:[type]
	PropertyHash uint32   `json:"propertyHash,omitempty"` // A hash used to determine if the properties have changed
}

func NewShape(keyNames, properties []string) (Shape, error) {

	shape := Shape{}
	keyHash := uint32(0)

	if keyNames != nil {
		var err error
		keyHash, err = hashArray(keyNames)
		if err != nil {
			return shape, err
		}
	}

	propHash, err := hashArray(properties)
	if err != nil {
		return shape, err
	}

	shape.KeyNames = keyNames
	shape.KeyNamesHash = keyHash
	shape.Properties = properties
	shape.PropertyHash = propHash
	return shape, nil

}

// EnsureHashes sets the hash values on the shape if they are unset.
func EnsureHashes(shape *Shape) {

	if shape.KeyNamesHash == 0 {
		shape.KeyNamesHash, _ = hashArray(shape.KeyNames)
	}

	if shape.PropertyHash == 0 {
		shape.PropertyHash, _ = hashArray(shape.Properties)
	}

}

// Shaper determines the schema of a given data point.  It will read through all the properties
// and return a Shape. This shape can be used to determine if the set of properties has changed
// between data points.
type Shaper interface {
	GetShape(keyNames []string, data map[string]interface{}) (Shape, error) // Gets the shape of a given data structure
}

type shaper struct {
}

type sortByPropName []string

func (s sortByPropName) Len() int {
	return len(s)
}

func (s sortByPropName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortByPropName) Less(i, j int) bool {
	s1Parts := strings.Split(s[i], ":")
	s2Parts := strings.Split(s[j], ":")

	s1Prop := strings.ToLower(s1Parts[0])
	s2Prop := strings.ToLower(s2Parts[0])

	return strings.Compare(s1Prop, s2Prop) < 0
}

// NewShaper creates a new instance of the default shaper.
func NewShaper() Shaper {
	return &shaper{}
}

func (s *shaper) GetShape(keyNames []string, data map[string]interface{}) (shape Shape, err error) {

	var properties []string

	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown error")
			}
			shape = Shape{}
		}
	}()

	getShapeRecursive(&properties, "", data)

	shape, err = NewShape(keyNames, properties)

	return shape, err
}

func getShapeRecursive(properties *[]string, prefix string, data map[string]interface{}) {

	for key, val := range data {

		if strings.Contains(key, ":") {
			panic("Invalid character found in property '" + key + "'.")
		}

		propName := getPropertyName(key, prefix)

		switch x := val.(type) {
		case string:
			if len(x) > 0 && unicode.IsDigit(rune(x[0])) && isDate(x) {
				*properties = append(*properties, propName+":date")
			} else {
				*properties = append(*properties, propName+":string")
			}
		case int, int8, int16, int32, int64, float32, float64:
			*properties = append(*properties, propName+":number")
		case bool:
			*properties = append(*properties, propName+":bool")
		case map[string]interface{}:
			*properties = append(*properties, propName+":object")
			getShapeRecursive(properties, propName, val.(map[string]interface{}))
		}
	}

}

func getPropertyName(name string, prefix string) string {
	propName := name
	if prefix != "" {
		propName = prefix + "." + propName
	}

	return propName
}

func isDate(val string) bool {

	if _, err := time.Parse(time.RFC3339, val); err == nil {
		return true
	}

	if _, err := time.Parse(time.RFC3339Nano, val); err == nil {
		return true
	}

	if _, err := time.Parse(time.RFC822, val); err == nil {
		return true
	}

	if _, err := time.Parse(time.RFC822Z, val); err == nil {
		return true
	}

	return false
}

func hashArray(properties []string) (uint32, error) {
	sort.Sort(sortByPropName(properties))

	// We are using a CRC check sum because it is very
	// efficient.  We are simply looking for a change,
	// we are not giving an identity.  Therefore, we don't
	// have to be concered about collisions.
	crcStr := ""
	propLen := len(properties)

	// We are using a lower case value for the properties
	// in order to allow for case in-sensitivity.
	for i, prop := range properties {
		crcStr = crcStr + strings.ToLower(prop)

		if i < (propLen - 1) {
			crcStr = crcStr + ","
		}
	}

	crc := crc32.New(castagnoliTable)

	if _, err := crc.Write([]byte(crcStr)); err != nil {
		return 0, err
	}

	return crc.Sum32(), nil
}

// ShapeDefinitions is a mapping of shape definition data
type ShapeDefinitions []ShapeDefinition

type ShapeDefinition struct {
	ID          string               `json:"id" bson:"_id,omitempty"`    // The ID of the ShapeDefinition
	Namespace   string               `json:"namespace,omitempty"`        // The namespace the shape definition belongs to
	Name        string               `json:"name,omitempty" bson:"name"` // The name of the shape definition (unique within Namespoce.)
	Description string               `json:"description,omitempty" bson:"description"`
	Keys        []string             `json:"keys,omitempty" bson:"keys"`
	Properties  []PropertyDefinition `json:"properties,omitempty" bson:"properties"`
}

// SortPropertyDefinitionsByName is an alias for []PropertyDefinition which implements sort.Interface using PropertyDefinition.Name.
type SortPropertyDefinitionsByName []PropertyDefinition

func (p SortPropertyDefinitionsByName) Len() int      { return len(p) }
func (p SortPropertyDefinitionsByName) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p SortPropertyDefinitionsByName) Less(i, j int) bool {
	return strings.Compare(p[i].Name, p[j].Name) < 0
}

type PropertyDefinition struct {
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
	Type        string `json:"type" bson:"type"`
}

type SortShapesByName ShapeDefinitions

func (s SortShapesByName) Len() int {
	return len(s)
}

func (s SortShapesByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s SortShapesByName) Less(i, j int) bool {
	s1Name := s[i].Name
	s2Name := s[j].Name
	return strings.Compare(s1Name, s2Name) < 0
}
