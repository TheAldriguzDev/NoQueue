package queue

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"

	"github.com/TheAldriguzDev/aldriguz_aio/packages/log"

	"github.com/disintegration/imaging"
)

// Image used to decode the queue it captcha
func decodeImage(base64Str string) (image.Image, error) {
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return img, nil
}

func SolveCaptcha(captchaImg string) (string, error) {
	img, err := decodeBase64Image(captchaImg)
	if err != nil {
		return "", err
	}

	// First we need to classify the captcha
	captchaType := classifyCaptcha(img)

	switch captchaType {
	case "Box captcha":

		// Solve box captcha

		// Target color represent the text color
		targetColor1 := color.RGBA{52, 66, 127, 255}
		targetColor2 := color.RGBA{16, 32, 107, 255}

		// We extract the color from the image (text color is +- blue / violet)
		outputImg1 := extractColor(img, targetColor1)
		outputImg2 := extractColor(img, targetColor2)

		// We can now combine the image
		combinedImg := combineImagesWithoutWhite(outputImg1, outputImg2)

		// We put the image to gray scale
		combinedImgGray := convertToGrayScale(combinedImg)

		// Color that we use to fill the gaps
		fillColor := color.RGBA{R: 0, G: 0, B: 0, A: 255}

		// We can now fill the gaps
		fill_gaps_image := fill_gaps(combinedImgGray, fillColor)

		//Set all pixel to black
		final_image := setNonWhitePixelsToBlack(fill_gaps_image)

		// We can now extract the text from the image
		text, err := getTextFromImage(final_image)

		if err != nil {
			return "", err
		}

		return text, nil
	case "Noise captcha":

		// We convert the image to grayscale
		grayImg := convertToGrayScale(img)

		// We can now reduce the noise
		img, err = reduceNoise(grayImg)

		// We can now get the text
		captchaText, err := getTextFromImage(img)

		if err != nil {
			return "", err
		}

		return captchaText, nil
	default:
		log.Error("No captcha has been found!", location)
		break
	}

	return "", nil
}

// Functions to classify the captcha
func classifyCaptcha(img image.Image) string {
	if VerifyCaptchaColor(img) {
		return "Box captcha"
	} else {
		return "Noise captcha"
	}
}

// Calculate de difference
func absDiff(a, b int) int {
	if a > b {
		return a - b
	}
	return b - a
}

// With this function we are going to classify the image:
/*
Idea: Since the queue it captcha follow 3 patterns:
1 - Box captcha
2 - Light noise captcha
3 - Dark noise captcha

We need to check if the top-left corner is white --> Box captcha has been found

Box captcha is the hardest one
*/
func VerifyCaptchaColor(img image.Image) bool {
	// Pixel coordinates used to classify the captcha
	x := 10
	y := 10

	// We get the pixels colors
	pixelColor := img.At(x, y)

	// Extract rgb colors
	r, g, b, _ := pixelColor.RGBA()

	// Verify if the color is in the correct range
	targetColorValue := 253
	colorTolerance := 30 // Color tollerance

	return absDiff(int(r>>8), targetColorValue) <= colorTolerance &&
		absDiff(int(g>>8), targetColorValue) <= colorTolerance &&
		absDiff(int(b>>8), targetColorValue) <= colorTolerance
}

// CASE BOX CAPTCHA
const colorTolerance = 50

func colorDistance(c1, c2 color.RGBA) int {
	r1, g1, b1 := c1.R, c1.G, c1.B
	r2, g2, b2 := c2.R, c2.G, c2.B

	rd := int(r1) - int(r2)
	gd := int(g1) - int(g2)
	bd := int(b1) - int(b2)

	return rd*rd + gd*gd + bd*bd
}

func extractColor(inputImg image.Image, targetColor color.RGBA) image.Image {
	bounds := inputImg.Bounds()
	outputImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixelColor := color.RGBAModel.Convert(inputImg.At(x, y)).(color.RGBA)
			if colorDistance(pixelColor, targetColor) <= colorTolerance*colorTolerance {
				outputImg.Set(x, y, pixelColor)
			} else {
				outputImg.Set(x, y, color.RGBA{255, 255, 255, 255}) // Set pixel to white
			}
		}
	}

	return outputImg
}

func combineImagesWithoutWhite(baseImg, overlayImg image.Image) image.Image {
	bounds := baseImg.Bounds()
	combinedImg := image.NewRGBA(bounds)

	draw.Draw(combinedImg, bounds, baseImg, image.Point{}, draw.Over)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixelColor := color.RGBAModel.Convert(overlayImg.At(x, y)).(color.RGBA)
			if pixelColor.R != 255 || pixelColor.G != 255 || pixelColor.B != 255 {
				combinedImg.Set(x, y, pixelColor)
			}
		}
	}

	return combinedImg
}

const whiteColorThreshold = 65025

func fill_gaps(img image.Image, fillColor color.Color) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	// Copy the original image in the new image
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			newImg.Set(x, y, img.At(x, y))
		}
	}

	// Fill the holes in the text
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// If the pixel is white, we try to fill it
			r, g, b, _ := img.At(x, y).RGBA()
			if r == 0xFFFF && g == 0xFFFF && b == 0xFFFF {
				// Check the near pixels
				neighbours := []image.Point{
					{X: x - 1, Y: y},
					{X: x + 1, Y: y},
					{X: x, Y: y - 1},
					{X: x, Y: y + 1},
				}

				// Search for a near colored pixel, if found fill it
				for _, p := range neighbours {
					if p.In(bounds) {
						r, g, b, _ := img.At(p.X, p.Y).RGBA()
						if r != 0xFFFF || g != 0xFFFF || b != 0xFFFF {
							newImg.Set(x, y, fillColor)
							break
						}
					}
				}
			}
		}
	}

	return newImg
}

func setNonWhitePixelsToBlack(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := img.At(x, y)

			// Check if the pixel is not white
			r, g, b, _ := pixel.RGBA()
			if r != 0xFFFF || g != 0xFFFF || b != 0xFFFF {
				// Set the pixel to black
				newImg.Set(x, y, color.RGBA{0, 0, 0, 255})
			} else {
				// Copy the original white pixel
				newImg.Set(x, y, pixel)
			}
		}
	}

	return newImg
}

// NOISE CAPTCHA
/*
We just need to process a bit the image
*/
func reduceNoise(img image.Image) (image.Image, error) {
	bounds := img.Bounds()
	dx, dy := bounds.Max.X, bounds.Max.Y

	// New image, where we save the manipulated one
	denoisedImg := imaging.New(dx, dy, color.Black)

	// Scan the image and apply a median filter to the pixels
	for y := 1; y < dy-1; y++ {
		for x := 1; x < dx-1; x++ {
			var r, g, b uint32
			var count uint32

			// Scan the surrounding pixels to calculate the average
			for j := -1; j <= 1; j++ {
				for i := -1; i <= 1; i++ {
					pixel := img.At(x+i, y+j)
					gray := color.GrayModel.Convert(pixel).(color.Gray)
					r += uint32(gray.Y)
					g += uint32(gray.Y)
					b += uint32(gray.Y)
					count++
				}
			}

			// Calculates the average value of the surrounding pixels
			avg := color.RGBA{
				R: uint8(r / count),
				G: uint8(g / count),
				B: uint8(b / count),
				A: 255,
			}

			// Set the pixel in the new image
			denoisedImg.Set(x, y, avg)
		}
	}

	width := denoisedImg.Bounds().Dx()
	cutWidth := int(float64(width) * 0.2)
	croppedImg := imaging.CropCenter(denoisedImg, width-cutWidth*2, denoisedImg.Bounds().Dy())

	finalImage := imaging.AdjustContrast(croppedImg, 60)

	finalImage = imaging.Sharpen(finalImage, 2.0)

	return finalImage, nil
}

func convertToGrayScale(img image.Image) *image.Gray {
	grayImage := image.NewGray(img.Bounds())
	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			grayImage.Set(x, y, color.GrayModel.Convert(img.At(x, y)))
		}
	}
	return grayImage
}
