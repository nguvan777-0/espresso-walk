package main

import (
	crand "crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"math"
	"math/rand"
)

// unitVector scales a 3D vector so its length is exactly 1.0
func unitVector(x, y, z float64) (float64, float64, float64) {
	length := math.Sqrt(x*x + y*y + z*z)
	if length == 0 {
		return 1, 0, 0
	}
	return x / length, y / length, z / length
}

// rotate3D uses Rodrigues' rotation formula to rotate a point around a 3D axis
func rotate3D(cx, cy, cz, px, py, pz, nx, ny, nz, angleDeg float64) (float64, float64, float64) {
	px, py, pz = px-cx, py-cy, pz-cz
	nx, ny, nz = unitVector(nx, ny, nz)

	rad := angleDeg * math.Pi / 180.0
	cos := math.Cos(rad)
	sin := math.Sin(rad)

	dot := px*nx + py*ny + pz*nz
	crossX := ny*pz - nz*py
	crossY := nz*px - nx*pz
	crossZ := nx*py - ny*px

	rx := px*cos + crossX*sin + nx*dot*(1-cos)
	ry := py*cos + crossY*sin + ny*dot*(1-cos)
	rz := pz*cos + crossZ*sin + nz*dot*(1-cos)

	return rx + cx, ry + cy, rz + cz
}

// Convert CIELAB to RGB
func labToRgb(l, a, b float64) (uint8, uint8, uint8) {
	y := (l + 16) / 116
	x := a/500 + y
	z := y - b/200

	f := func(t float64) float64 {
		if math.Pow(t, 3) > 0.008856 {
			return math.Pow(t, 3)
		}
		return (t - 16.0/116.0) / 7.787
	}

	x = 0.95047 * f(x)
	y = 1.00000 * f(y)
	z = 1.08883 * f(z)

	rLin := x*3.2406 + y*-1.5372 + z*-0.4986
	gLin := x*-0.9689 + y*1.8758 + z*0.0415
	bLin := x*0.0557 + y*-0.2040 + z*1.0570

	gamma := func(c float64) float64 {
		if c > 0.0031308 {
			return 1.055*math.Pow(c, 1/2.4) - 0.055
		}
		return 12.92 * c
	}

	R := math.Max(0, math.Min(1, gamma(rLin)))
	G := math.Max(0, math.Min(1, gamma(gLin)))
	B := math.Max(0, math.Min(1, gamma(bLin)))

	return uint8(R * 255), uint8(G * 255), uint8(B * 255)
}

func main() {
	stepsFlag := flag.Int("steps", 12, "Number of steps in the walk")
	gapFlag := flag.Bool("gap", false, "Add a newline gap between color bars")
    flag.Usage = func() {
        fmt.Println("Usage: espresso-walk [options]")
        fmt.Println("Runs a physics-based 3D orbital color walk anchored around espresso brown.")
        fmt.Println("\nOptions:")
        flag.PrintDefaults()
    }

	flag.Parse()

	var b [8]byte
	crand.Read(b[:])
	rng := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(b[:]))))

	// 1. Select a random espresso color variation
	targetL := 20.0 + rng.NormFloat64()*5.0
	targetA := 12.0 + rng.NormFloat64()*3.0
	targetB := 18.0 + rng.NormFloat64()*4.0

	// 2. Anchor the rotation point at 50% lightness (L=50) and neutral chromaticity (a=0, b=0).
	// The target espresso point is near L=20, giving a radius of 30-40 units.
	// Rotating around L=50 keeps the orbit between L=10 and L=90.
	centerL, centerA, centerB := 50.0, 0.0, 0.0

	// 3. Select a random 3D rotation axis
	axisL := rng.NormFloat64()
	axisA := rng.NormFloat64()
	axisB := rng.NormFloat64()

	anchorIndex := rng.Intn(*stepsFlag)
	degreesPerStep := 360.0 / float64(*stepsFlag)

	// The target color is at the randomly chosen anchorIndex.
	// Calculate the starting point (step 0) by reversing the rotation from the anchor point.
	startAngle := -float64(anchorIndex) * degreesPerStep
	startL, startA, startB := rotate3D(centerL, centerA, centerB, targetL, targetA, targetB, axisL, axisA, axisB, startAngle)

	fmt.Printf("\nespresso-walk\n\n")

	for i := 0; i < *stepsFlag; i++ {
		angle := float64(i) * degreesPerStep

		currL, currA, currB := rotate3D(centerL, centerA, centerB, startL, startA, startB, axisL, axisA, axisB, angle)

		r, g, b := labToRgb(currL, currA, currB)

		colorCode := fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
		resetCode := "\x1b[0m"

		fmt.Printf("%s                                %s\n", colorCode, resetCode)

		if *gapFlag && i < *stepsFlag-1 {
			fmt.Printf("\n")
		}
	}
	fmt.Printf("\n")
}
