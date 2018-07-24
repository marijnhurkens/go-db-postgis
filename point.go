package postgis

import (
	"bytes"
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
)

// Point represents a Postgis POINT
type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// String casts to string
func (p *Point) String() string {
	return fmt.Sprintf("SRID=4326;POINT(%v %v)", p.Lng, p.Lat)
}

// Scan decodes a bytearray to the point structure
func (p *Point) Scan(val interface{}) error {
	// decode from hex string to byte array
	b, err := hex.DecodeString(string(val.([]uint8)))
	if err != nil {
		return err
	}

	// Read the byteorder
	r := bytes.NewReader(b)
	var wkbByteOrder uint8
	if err := binary.Read(r, binary.LittleEndian, &wkbByteOrder); err != nil {
		return err
	}

	var byteOrder binary.ByteOrder
	switch wkbByteOrder {
	case 0:
		byteOrder = binary.BigEndian
	case 1:
		byteOrder = binary.LittleEndian
	default:
		return fmt.Errorf("Invalid byte order %d", wkbByteOrder)
	}

	// Read geometry type
	var wkbGeometryType uint64
	if err := binary.Read(r, byteOrder, &wkbGeometryType); err != nil {
		return err
	}

	// Read the lat / lng into the point structure
	if err := binary.Read(r, byteOrder, p); err != nil {
		return err
	}

	// No error so return nil
	return nil
}

// Value casts to string for the driver
func (p Point) Value() (driver.Value, error) {
	return p.String(), nil
}

// MarshalJSON implements the JSON marshaller for the nullstring
func (p Point) MarshalJSON() ([]byte, error) {

	bytes := bytes.NewBufferString("{\"lat\":" + strconv.FormatFloat(p.Lat, 'f', -1, 64) + ",\"lng\":" + strconv.FormatFloat(p.Lng, 'f', -1, 64) + "}")

	return bytes.Bytes(), nil

}

// NullPoint is used to create a nullable point
type NullPoint struct {
	Point Point
	Valid bool
}

// Scan checks if bytearray exists and decodes if neccessary
func (np *NullPoint) Scan(val interface{}) error {
	// If val is nil create a nullpoint with an empty point
	if val == nil {
		np.Point, np.Valid = Point{}, false
		return nil
	}

	// Init a new point and scan the val into the point
	point := &Point{}
	err := point.Scan(val)
	if err != nil {
		np.Point, np.Valid = Point{}, false
		return nil
	}

	// Fill the data
	np.Point = Point{
		Lat: point.Lat,
		Lng: point.Lng,
	}
	np.Valid = true

	return nil
}

// Value returns nil of empty or casts to string
func (np NullPoint) Value() (driver.Value, error) {
	if !np.Valid {
		return nil, nil
	}
	return np.Point.Value()
}

// MarshalJSON implements the JSON marshaller for the nullpoint
func (np NullPoint) MarshalJSON() ([]byte, error) {
	if np.Valid {
		bytes := bytes.NewBufferString("{\"lat\":" + strconv.FormatFloat(np.Point.Lat, 'f', -1, 64) + ",\"lng\":" + strconv.FormatFloat(np.Point.Lng, 'f', -1, 64) + "}")

		return bytes.Bytes(), nil
	}

	bytes := bytes.NewBufferString("null")

	return bytes.Bytes(), nil
}

// UnmarshalJSON implements the JSON unmarshaller for the nullpoint
func (np *NullPoint) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &np.Point)
	np.Valid = (err == nil)
	return err
}
