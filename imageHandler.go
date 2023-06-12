package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"sync"

	"gonum.org/v1/gonum/mat"
)

type ImageHandler struct{}

func NewImageHandler() *ImageHandler {
	return &ImageHandler{}
}

type Tensor = [][]color.Color

type BufferSize struct {
	X int
	Y int
}

func (handler *ImageHandler) CreateTensor(img image.Image) Tensor {
	size := img.Bounds().Size()

	var pixels Tensor

	for i := 0; i < size.X; i++ {
		var y []color.Color
		for j := 0; j < size.Y; j++ {
			y = append(y, img.At(i, j))
		}
		pixels = append(pixels, y)
	}

	return pixels
}

func (handler *ImageHandler) OpenImage(path string) (image.Image, error) {
	file, err := os.Open(path)

	if err != nil {
		fmt.Println("File could not be loaded.")

		return nil, err
	}

	fileInfo, _ := file.Stat()

	fmt.Printf("Loaded file: %s \n", fileInfo.Name())

	img, _, err := image.Decode(file)

	return img, err
}

func (handler *ImageHandler) GrayScaleImage(imgBuffer *Tensor, wg *sync.WaitGroup, mu *sync.Mutex) {
	mu.Lock()
	bufferSize := BufferSize{
		X: len(*imgBuffer),
		Y: len((*imgBuffer)[0]),
	}

	for x := 0; x < bufferSize.X; x++ {
		for y := 0; y < bufferSize.Y; y++ {
			originalColor, ok := color.RGBAModel.Convert((*imgBuffer)[x][y]).(color.RGBA)

			if !ok {
				fmt.Println("Color conversion failed.")
			}

			grey := uint8(float64(originalColor.G)*0.72 + float64(originalColor.B)*0.07 + float64(originalColor.R)*0.21)
			(*imgBuffer)[x][y] = color.RGBA{grey, grey, grey, originalColor.A}
		}
	}
	mu.Unlock()
	wg.Done()
}

func (handler *ImageHandler) RotateImage(imgBuffer *Tensor, wg *sync.WaitGroup, mu *sync.Mutex) {
	mu.Lock()
	bufferSize := BufferSize{
		X: len(*imgBuffer),
		Y: len((*imgBuffer)[0]),
	}

	for x := 0; x < bufferSize.X; x++ {
		for y := 0; y < bufferSize.Y/2; y++ {
			(*imgBuffer)[x][y], (*imgBuffer)[bufferSize.X-x-1][bufferSize.Y-y-1] = (*imgBuffer)[bufferSize.X-x-1][bufferSize.Y-y-1], (*imgBuffer)[x][y]
		}
	}
	mu.Unlock()
	wg.Done()
}

func (handler *ImageHandler) BlurImage(imgBuffer *Tensor, kernel *mat.Dense, wg *sync.WaitGroup, mu *sync.Mutex) {
	rows, cols := kernel.Dims()
	offset := rows / 2
	kernelLength := cols

	mu.Lock()
	for x := 0; x < len(*imgBuffer); x++ {
		for y := 0; y < len((*imgBuffer)[0]); y++ {
			newPixel := color.RGBA{}

			for a := 0; a < kernelLength; a++ {
				for b := 0; b < kernelLength; b++ {
					xn := math.Max(math.Min(float64(x+a-offset), float64(len(*imgBuffer)-1)), 0)
					yn := math.Max(math.Min(float64(y+b-offset), float64(len((*imgBuffer)[0])-1)), 0)

					r, g, bb, aa := (*imgBuffer)[int(xn)][int(yn)].RGBA()

					newPixel.R += uint8(float64(uint8(r)) * kernel.At(a, b))
					newPixel.G += uint8(float64(uint8(g)) * kernel.At(a, b))
					newPixel.B += uint8(float64(uint8(bb)) * kernel.At(a, b))
					newPixel.A += uint8(float64(uint8(aa)) * kernel.At(a, b))
				}
			}
			(*imgBuffer)[x][y] = newPixel
		}
	}
	mu.Unlock()
	wg.Done()
}

func (handler *ImageHandler) DecodeTensor(pixels [][]color.Color) image.Image {
	rect := image.Rect(0, 0, len(pixels), len(pixels[0]))
	newImage := image.NewRGBA(rect)

	for x := 0; x < len(pixels); x++ {
		for y := 0; y < len(pixels[0]); y++ {
			q := pixels[x]

			if q == nil {
				continue
			}

			p := pixels[x][y]

			if p == nil {
				continue
			}

			original, ok := color.RGBAModel.Convert(p).(color.RGBA)

			if ok {
				newImage.Set(x, y, original)
			}
		}
	}
	return newImage
}
