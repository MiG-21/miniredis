package miniredis

import (
	"math"

	"github.com/alicebob/miniredis/v2/geohash"
)

// EarthRadius is the radius of the earth in meters. It is used in geo distance calculations.
// To keep things consistent, this value matches WGS84 Web Mercator (EPSG:3857).
const EarthRadius = 6372797.560856 // meters

func deg2rad(d float64) float64 {
	return d * math.Pi / 180.0
}

func rad2deg(r float64) float64 {
	return 180.0 * r / math.Pi
}

func toGeohash(long, lat float64) uint64 {
	return geohash.EncodeIntWithPrecision(lat, long, 52)
}

func fromGeohash(score uint64) (float64, float64) {
	lat, long := geohash.DecodeIntWithPrecision(score, 52)
	return long, lat
}

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// distance function returns the distance (in meters) between two points of
// a given longitude and latitude relatively accurately (using a spherical
// approximation of the Earth) through the Haversin Distance Formula for
// great arc distance on a sphere with accuracy for small distances
// point coordinates are supplied in degrees and converted into rad. in the func
// distance returned is meters
// http://en.wikipedia.org/wiki/Haversine_formula
// Source: https://gist.github.com/cdipaolo/d3f8db3848278b49db68
func distance(lat1, lon1, lat2, lon2 float64) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2 float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	earth := 6372797.560856 // Earth radius in METERS, according to src/geohash_helper.c

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * earth * math.Asin(math.Sqrt(h))
}

/* Judge whether a point is in the axis-aligned rectangle, when the distance
 * between a searched point and the center point is less than or equal to
 * height/2 or width/2 in height and width, the point is in the rectangle.
 *
 * width_m, height_m: the rectangle
 * x1, y1 : the center of the box
 * x2, y2 : the point to be searched
 */
func geohashGetDistanceIfInRectangle(widthM, heightM, x1, y1, x2, y2 float64) (float64, bool) {
	/* latitude distance is less expensive to compute than longitude distance
	 * so we check first for the latitude condition */
	latDistance := geohashGetLatDistance(y2, y1)
	if latDistance > heightM/2 {
		return 0, false
	}
	lonDistance := geohashGetDistance(x2, y2, x1, y2)
	if lonDistance > widthM/2 {
		return 0, false
	}

	return geohashGetDistance(x1, y1, x2, y2), true
}

/* Calculate distance using simplified haversine great circle distance formula.
 * Given longitude diff is 0 the asin(sqrt(a)) on the haversine is asin(sin(abs(u))).
 * arcsin(sin(x)) equal to x when x ∈[−𝜋/2,𝜋/2]. Given latitude is between [−𝜋/2,𝜋/2]
 * we can simplify arcsin(sin(x)) to x.
 */
func geohashGetLatDistance(lat1d, lat2d float64) float64 {
	return EarthRadius * math.Abs(deg2rad(lat2d)-deg2rad(lat1d))
}

/* Calculate distance using haversine great circle distance formula. */
func geohashGetDistance(lon1d, lat1d, lon2d, lat2d float64) float64 {
	lon1r := deg2rad(lon1d)
	lon2r := deg2rad(lon2d)
	v := math.Sin((lon2r - lon1r) / 2)
	/* if v == 0 we can avoid doing expensive math when lons are practically the same */
	if v == 0 {
		return geohashGetLatDistance(lat1d, lat2d)
	}
	lat1r := deg2rad(lat1d)
	lat2r := deg2rad(lat2d)
	u := math.Sin((lat2r - lat1r) / 2)
	a := u*u + math.Cos(lat1r)*math.Cos(lat2r)*v*v

	return 2.0 * EarthRadius * math.Asin(math.Sqrt(a))
}
