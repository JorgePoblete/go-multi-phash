package main

import (
	"bytes"
	"encoding/hex"
	"image"
	"image/jpeg"
	"log"
	"math"
	"os"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/nfnt/resize"
)

type Image struct {
	img image.Image
}

func (i *Image) Resize(w, h uint) *Image {
	i.img = resize.Resize(w, h, i.img, resize.Lanczos3)
	return i
}

func (i *Image) GrayScale() *Image {
	i.img = imaging.Grayscale(i.img)
	return i
}

func (i *Image) Copy() *Image {
	return &Image{
		img: i.img,
	}
}

func (i *Image) SubImage(x0, y0, x1, y1 int) *Image {
	type SubImageType interface {
		SubImage(r image.Rectangle) image.Image
	}
	return &Image{
		img: i.img.(SubImageType).SubImage(image.Rect(x0, y0, x1, y1)),
	}
}

func (i *Image) Float64() []float64 {
	float64Image := []float64{}
	bounds := i.img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := i.img.At(x, y).RGBA()
			rgb := math.Sqrt(float64(r)) + math.Sqrt(float64(g)) + math.Sqrt(float64(b))
			float64Image = append(float64Image, rgb/3)
		}
	}
	return float64Image
}

func (i *Image) Bytes() []byte {
	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, i.img, nil)
	if err != nil {
		log.Printf("There was an error decoding the image: %+v", err)
	}
	return buf.Bytes()
}

func (i *Image) Signature() string {
	gridSize := 11
	neighbors := 8
	bounds := i.img.Bounds()
	xPadding := (bounds.Max.X / gridSize) - 1
	yPadding := (bounds.Max.Y / gridSize) - 1
	compute := Compute{}
	means := []float64{}
	grayImg := i.Copy().GrayScale()
	// now compute each grid block and each grid block mean
	// the center of each block of pixels should be
	// (6,6), (6,12), (6,18), (6,24), (6,30), (6,36), (6,42), (6,48), (6,54)
	// (12,6),(12,12),(12,18),(12,24),(12,30),(12,36),(12,42),(12,48),(12,54)
	// (18,6),(18,18),(18,18),(18,24),(18,30),(18,36),(18,42),(18,48),(18,54)
	for k := 0; k < gridSize; k++ {
		y := (k * yPadding) + yPadding
		for m := 0; m < gridSize; m++ {
			x := (m * xPadding) + xPadding
			// get the block of the <neighbors> surrounding pixels
			block := grayImg.SubImage(x-neighbors, y-neighbors, x+neighbors, y+neighbors)
			// then compute the block mean value
			// and use it as the block value for next step
			mean := compute.Mean(block.Float64())
			means = append(means, mean)
		}
	}
	// for simplicity of the next step i will normalize the data to a -2,2 range
	// knowing this is a grayscale image i have than min and max values will be 0 and 255
	// x' = (max'-min')/(max-min)*(x-max)+max'
	a := (2.0 - (-2.0)) / (255.0 - 0.0)
	for j, x := range means {
		means[j] = math.Round((a * (x - 255.0)) + 2.0)
	}
	// now using the means compute the signature vectors based on the neighbors of
	// each of the gridSize*gridSize means
	// we should have a maxium of 8-neighbors for each mean, so a maxium of
	// 8 comparations can be made
	signatureVector := []int{}
	// this will be an ugly if zone
	for y := 0; y < gridSize; y++ {
		for x := 0; x < gridSize; x++ {
			a := means[(y*gridSize)+x]
			// topleft
			if y == 0 {
				signatureVector = append(signatureVector, 0)
			} else if x == 0 {
				signatureVector = append(signatureVector, 0)
			} else {
				value := means[((y-1)*gridSize)+x-1] - a
				signatureVector = append(signatureVector, int(value))
			}
			// top
			if y == 0 {
				signatureVector = append(signatureVector, 0)
			} else {
				value := means[((y-1)*gridSize)+x] - a
				signatureVector = append(signatureVector, int(value))
			}
			// topright
			if y == 0 {
				signatureVector = append(signatureVector, 0)
			} else if x == gridSize-1 {
				signatureVector = append(signatureVector, 0)
			} else {
				value := means[((y-1)*gridSize)+x+1] - a
				signatureVector = append(signatureVector, int(value))
			}
			// left
			if x == 0 {
				signatureVector = append(signatureVector, 0)
			} else {
				value := means[(y*gridSize)+x-1] - a
				signatureVector = append(signatureVector, int(value))
			}
			// right
			if x == gridSize-1 {
				signatureVector = append(signatureVector, 0)
			} else {
				value := means[(y*gridSize)+x+1] - a
				signatureVector = append(signatureVector, int(value))
			}
			// bottomleft
			if y == gridSize-1 {
				signatureVector = append(signatureVector, 0)
			} else if x == 0 {
				signatureVector = append(signatureVector, 0)
			} else {
				value := means[((y+1)*gridSize)+x-1] - a
				signatureVector = append(signatureVector, int(value))
			}
			// bottom
			if y == gridSize-1 {
				signatureVector = append(signatureVector, 0)
			} else {
				value := means[((y+1)*gridSize)+x] - a
				signatureVector = append(signatureVector, int(value))
			}
			// bottomright
			if y == gridSize-1 {
				signatureVector = append(signatureVector, 0)
			} else if x == gridSize-1 {
				signatureVector = append(signatureVector, 0)
			} else {
				value := means[((y+1)*gridSize)+x+1] - a
				signatureVector = append(signatureVector, int(value))
			}
		}
	}
	// normalize the [-4,4] data to a [-2,2] range
	// to make it fit a darker, much darker or lighter, much lighter scheme
	// this will make some data loss but meh
	for j, x := range signatureVector {
		if x < -2 {
			signatureVector[j] = -2
		}
		if x > 2 {
			signatureVector[j] = 2
		}
	}
	// until here we already have something that can be called a signature
	// but i will convert them into a series of power of 3 numbers that
	// can be easily indexed on the db
	// first i have to simplify the [-2,2] data to a [-1,1] range
	// this will make some data loss but meh
	// then i will add 1 to each number so the range will now be [0-2]
	// this gives 3 posible values 0, 1, 2
	// this data will then be choped into 63 overlaping chunks of 16 values each
	// and will be dot multiplied with a vector with power of 3 values
	// to obtain the value that will be used as an index
	for j, x := range signatureVector {
		value := signatureVector[j]
		if x == -2 {
			value = -1
		}
		if x == 2 {
			value = 1
		}
		signatureVector[j] = value + 1
	}

	powerOf3Values := []int{1, 3, 9, 27, 81, 243, 729, 2187, 6561, 19683, 59049, 177147, 531441, 1594323, 4782969, 14348907}
	position := 0
	indexes := []int{}
	// there will be 63 values
	for j := 0; j < 63; j++ {
		// im making chunks of 16 values each
		indexes = append(
			indexes,
			compute.IntegerDotMultiplication(
				powerOf3Values,
				signatureVector[position:position+16],
			),
		)
		// and incrementing the position by ten, so the chunks overlap in 6 values
		position += 10
	}
	result := []string{}
	for _, x := range indexes {
		result = append(result, string(x))
	}
	stringResult := strings.Join(result, "")
	return hex.EncodeToString([]byte(stringResult))
	/*
		return indexes
	*/
}

func NewImage(img string) *Image {
	file, err := os.Open(img)
	if err != nil {
		return &Image{}
	}
	defer file.Close()

	image, _, err := image.Decode(file)
	if err != nil {
		return &Image{}
	}
	return &Image{
		img: image,
	}
}
