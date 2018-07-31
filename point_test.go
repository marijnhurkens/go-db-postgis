package postgis_test

import (
	"github.com/marijnhurkens/go-db-postgis"
	"testing"
)

func TestPointToString(t *testing.T) {
	point := postgis.Point{
		Lat: 10.5,
		Lng: 2.3,
	}

	expected := "SRID=4326;POINT(2.3 10.5)"

	if point.String() != expected {
		t.Error("Point to string not expected string, exdpected:", expected, ", actual:", point.String())
	}
}

func TestPointScan(t *testing.T) {
	var point postgis.Point

	// lat: 1 , lng: 2 encoded as ewkb hex
	data := "0101000020E51000000000000000000040000000000000F03F"

	point.Scan([]uint8(data))

	if point.Lat != 1 {
		t.Error("Lat not 1, actual:", point.Lat)
	}

	if point.Lng != 2 {
		t.Error("Lng not 2, actual:", point.Lng)
	}
}
